package svgjsx

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

func parseDimension(v string) (int, bool) {
	v = strings.TrimSpace(v)
	v = strings.TrimSuffix(v, "px")
	n, err := strconv.Atoi(v)
	if err != nil || n < 0 {
		return 0, false
	}
	return n, true
}

func dimensionClass(prefix string, px int) string {
	if px%4 == 0 {
		return fmt.Sprintf("%s-%d", prefix, px/4)
	}
	return fmt.Sprintf("%s-[%dpx]", prefix, px)
}

func tailwindSize(wStr, hStr string) (classes []string, keptW, keptH string) {
	wPx, wOk := parseDimension(wStr)
	hPx, hOk := parseDimension(hStr)

	if wStr != "" && hStr != "" && wOk && hOk {
		if wPx == hPx {
			classes = append(classes, dimensionClass("size", wPx))
		} else {
			classes = append(classes, dimensionClass("w", wPx), dimensionClass("h", hPx))
		}
		return classes, "", ""
	}

	if wStr != "" {
		if wOk {
			classes = append(classes, dimensionClass("w", wPx))
		} else {
			keptW = wStr
		}
	}
	if hStr != "" {
		if hOk {
			classes = append(classes, dimensionClass("h", hPx))
		} else {
			keptH = hStr
		}
	}
	return classes, keptW, keptH
}

func tailwindColor(prefix, val string) (cls string, kept bool) {
	v := strings.TrimSpace(val)
	if strings.HasPrefix(strings.ToLower(v), "url(") {
		return "", true
	}
	lower := strings.ToLower(v)
	switch lower {
	case "none":
		return prefix + "-none", false
	case "currentcolor", "current":
		return prefix + "-current", false
	case "transparent":
		return prefix + "-transparent", false
	case "black", "#000", "#000000":
		return prefix + "-black", false
	case "white", "#fff", "#ffffff":
		return prefix + "-white", false
	}

	if strings.HasPrefix(v, "#") {
		hex := strings.TrimPrefix(lower, "#")
		if len(hex) == 3 {
			hex = string([]byte{hex[0], hex[0], hex[1], hex[1], hex[2], hex[2]})
		}
		if shade, ok := tailwindColors[hex]; ok {
			return prefix + "-" + shade, false
		}
	}
	// Tailwind reads an underscore as a space inside arbitrary values.
	return prefix + "-[" + strings.ReplaceAll(v, " ", "_") + "]", false
}

func tailwindStrokeWidth(val string) string {
	v := strings.TrimSpace(val)
	v = strings.TrimSuffix(v, "px")
	switch v {
	case "0":
		return "stroke-0"
	case "1":
		return "stroke-1"
	case "2":
		return "stroke-2"
	default:
		return "stroke-[" + v + "]"
	}
}

func tailwindOpacity(val string) string {
	v := strings.TrimSpace(val)
	f, err := strconv.ParseFloat(v, 64)
	if err != nil {
		return "opacity-[" + v + "]"
	}

	pct := f * 100
	rounded := math.Round(pct)
	if math.Abs(pct-rounded) < 1e-4 {
		iPct := int(rounded)
		if iPct%5 == 0 && iPct >= 0 && iPct <= 100 {
			return fmt.Sprintf("opacity-%d", iPct)
		}
	}

	str := strconv.FormatFloat(f, 'f', -1, 64)
	if strings.HasPrefix(str, "0.") {
		str = str[1:]
	} else if strings.HasPrefix(str, "-0.") {
		str = "-" + str[2:]
	}
	return "opacity-[" + str + "]"
}
