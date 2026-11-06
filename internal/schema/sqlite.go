package schema

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

func readVarint(data []byte) (int64, int, error) {
	if len(data) == 0 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	var v uint64
	for i := 0; i < 8 && i < len(data); i++ {
		b := data[i]
		v = (v << 7) | uint64(b&0x7f)
		if (b & 0x80) == 0 {
			return int64(v), i + 1, nil
		}
	}
	if len(data) < 9 {
		return 0, 0, io.ErrUnexpectedEOF
	}
	v = (v << 8) | uint64(data[8])
	return int64(v), 9, nil
}

// leafPayload reassembles a table leaf cell, following overflow pages when needed.
// The size formulas come from the SQLite file format, section 1.6.
func leafPayload(cell []byte, usable, maxSize int, getPage func(uint32) ([]byte, error)) ([]byte, error) {
	size64, n1, err := readVarint(cell)
	if err != nil {
		return nil, err
	}
	if size64 < 0 || size64 > int64(maxSize) {
		return nil, fmt.Errorf("invalid payload size %d", size64)
	}
	size := int(size64)

	_, n2, err := readVarint(cell[n1:])
	if err != nil {
		return nil, err
	}
	offset := n1 + n2

	inline := size
	if size > usable-35 {
		minLocal := (usable-12)*32/255 - 23
		inline = minLocal + (size-minLocal)%(usable-4)
		if inline > usable-35 {
			inline = minLocal
		}
	}
	if offset+inline > len(cell) {
		return nil, errors.New("inline payload exceeds cell data")
	}
	payload := make([]byte, 0, size)
	payload = append(payload, cell[offset:offset+inline]...)
	if size == inline {
		return payload, nil
	}

	if offset+inline+4 > len(cell) {
		return nil, errors.New("overflow pointer out of bounds")
	}
	next := binary.BigEndian.Uint32(cell[offset+inline:])
	seen := map[uint32]bool{}
	for len(payload) < size {
		if next == 0 || seen[next] {
			return nil, errors.New("broken overflow chain")
		}
		seen[next] = true
		page, err := getPage(next)
		if err != nil {
			return nil, err
		}
		take := min(size-len(payload), usable-4)
		payload = append(payload, page[4:4+take]...)
		next = binary.BigEndian.Uint32(page)
	}
	return payload, nil
}

// serialSize returns the byte length of a record value.
func serialSize(st int64) (int, error) {
	switch {
	case st < 0 || st == 10 || st == 11:
		return 0, fmt.Errorf("unknown serial type %d", st)
	case st <= 4:
		return []int{0, 1, 2, 3, 4}[st], nil
	case st == 5:
		return 6, nil
	case st == 6 || st == 7:
		return 8, nil
	case st <= 9:
		return 0, nil
	}
	return int((st - 12) / 2), nil
}

// decodeRecord returns the text columns of a record, other types come back empty.
func decodeRecord(payload []byte) ([]string, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	headerSize, pos, err := readVarint(payload)
	if err != nil {
		return nil, err
	}
	if headerSize > int64(len(payload)) || headerSize < int64(pos) {
		return nil, fmt.Errorf("invalid record header size %d", headerSize)
	}

	var values []string
	data := int(headerSize)
	for pos < int(headerSize) {
		st, n, err := readVarint(payload[pos:int(headerSize)])
		if err != nil {
			return nil, err
		}
		pos += n
		size, err := serialSize(st)
		if err != nil {
			return nil, err
		}
		if size < 0 || data+size > len(payload) {
			return nil, io.ErrUnexpectedEOF
		}
		var v string
		if st >= 13 && st%2 == 1 {
			v = string(payload[data : data+size])
		}
		values = append(values, v)
		data += size
	}
	return values, nil
}

// ReadSQLite parses an SQLite database file directly and returns a Schema.
func ReadSQLite(path string) (*Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if len(data) < 100 {
		return nil, fmt.Errorf("file too small for sqlite")
	}
	if string(data[:16]) != "SQLite format 3\x00" {
		return nil, fmt.Errorf("not an sqlite file")
	}

	rawPageSize := binary.BigEndian.Uint16(data[16:18])
	var pageSize int
	if rawPageSize == 1 {
		pageSize = 65536
	} else {
		pageSize = int(rawPageSize)
	}
	if pageSize < 512 || pageSize > 65536 || (pageSize&(pageSize-1)) != 0 {
		return nil, fmt.Errorf("invalid page size: %d", pageSize)
	}

	reserved := int(data[20])
	U := pageSize - reserved
	if U < 480 {
		return nil, fmt.Errorf("invalid usable page size: %d", U)
	}

	totalPages := len(data) / pageSize
	if totalPages < 1 {
		return nil, fmt.Errorf("file has no full pages")
	}

	getPage := func(pageNum uint32) ([]byte, error) {
		if pageNum < 1 || int(pageNum) > totalPages {
			return nil, fmt.Errorf("page %d out of bounds (1..%d)", pageNum, totalPages)
		}
		offset := int(pageNum-1) * pageSize
		if offset+pageSize > len(data) {
			return nil, fmt.Errorf("page %d truncated", pageNum)
		}
		return data[offset : offset+pageSize], nil
	}

	visited := make(map[uint32]bool)
	maxPages := totalPages + 100
	var records [][]byte

	var visitPage func(pageNum uint32) error
	visitPage = func(pageNum uint32) error {
		if visited[pageNum] {
			return fmt.Errorf("cycle detected at page %d", pageNum)
		}
		visited[pageNum] = true
		if len(visited) > maxPages {
			return fmt.Errorf("too many pages visited")
		}

		page, err := getPage(pageNum)
		if err != nil {
			return err
		}

		hdrOffset := 0
		if pageNum == 1 {
			hdrOffset = 100
		}
		if len(page) < hdrOffset+8 {
			return fmt.Errorf("page %d too small for btree header", pageNum)
		}

		pageType := page[hdrOffset]
		numCells := int(binary.BigEndian.Uint16(page[hdrOffset+3 : hdrOffset+5]))

		switch pageType {
		case 0x05:
			if len(page) < hdrOffset+12 {
				return fmt.Errorf("interior page %d header truncated", pageNum)
			}
			rightChild := binary.BigEndian.Uint32(page[hdrOffset+8 : hdrOffset+12])
			cellPtrOffset := hdrOffset + 12

			for i := 0; i < numCells; i++ {
				if cellPtrOffset+2 > len(page) {
					return fmt.Errorf("cell pointer out of bounds on page %d", pageNum)
				}
				cellPtr := int(binary.BigEndian.Uint16(page[cellPtrOffset : cellPtrOffset+2]))
				cellPtrOffset += 2

				if cellPtr < 0 || cellPtr+4 > len(page) {
					return fmt.Errorf("cell offset %d out of bounds on page %d", cellPtr, pageNum)
				}
				leftChild := binary.BigEndian.Uint32(page[cellPtr : cellPtr+4])
				if err := visitPage(leftChild); err != nil {
					return err
				}
			}
			return visitPage(rightChild)

		case 0x0D:
			cellPtrOffset := hdrOffset + 8
			for i := 0; i < numCells; i++ {
				if cellPtrOffset+2 > len(page) {
					return fmt.Errorf("cell pointer out of bounds on page %d", pageNum)
				}
				cellPtr := int(binary.BigEndian.Uint16(page[cellPtrOffset : cellPtrOffset+2]))
				cellPtrOffset += 2

				if cellPtr < 0 || cellPtr >= len(page) {
					return fmt.Errorf("cell offset %d out of bounds on page %d", cellPtr, pageNum)
				}

				payload, err := leafPayload(page[cellPtr:], U, len(data), getPage)
				if err != nil {
					return err
				}
				records = append(records, payload)
			}
			return nil

		default:
			return fmt.Errorf("unexpected btree page type 0x%02x on page %d", pageType, pageNum)
		}
	}

	if err := visitPage(1); err != nil {
		return nil, err
	}

	var sqlParts []string
	for _, rec := range records {
		vals, err := decodeRecord(rec)
		if err != nil {
			return nil, err
		}
		if len(vals) < 5 {
			continue
		}
		typ, name, stmt := vals[0], vals[1], strings.TrimSpace(vals[4])
		if stmt == "" || strings.HasPrefix(name, "sqlite_") {
			continue
		}
		if typ == "table" || typ == "index" || typ == "view" {
			sqlParts = append(sqlParts, stmt)
		}
	}

	return ParseDDL(strings.Join(sqlParts, ";\n") + ";\n")
}
