package schema

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const pgDump = `--
-- PostgreSQL database dump
--
SET statement_timeout = 0;
SELECT pg_catalog.set_config('search_path', '', false);

CREATE TYPE public.order_status AS ENUM (
    'new',
    'paid'
);

CREATE FUNCTION public.touch() RETURNS trigger
    LANGUAGE plpgsql
    AS $$
BEGIN
  CREATE TABLE nope (id int);
  RETURN NEW;
END;
$$;

CREATE TABLE public.orgs (
    id bigint NOT NULL,
    name character varying(120) NOT NULL
);

CREATE TABLE public.users (
    id bigint NOT NULL,
    org_id bigint,
    email character varying(255) NOT NULL,
    balance numeric(10,2) DEFAULT 0.00 NOT NULL,
    meta jsonb DEFAULT '{"a": "b,c"}'::jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL
);

CREATE TABLE billing.invoices (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    user_id bigint NOT NULL,
    status public.order_status DEFAULT 'new'::public.order_status
);

COMMENT ON TABLE public.users IS 'Registered accounts';
COMMENT ON COLUMN public.users.email IS 'Login e-mail, it''s unique';

CREATE VIEW public.active_users AS
 SELECT users.id FROM public.users;

ALTER TABLE ONLY public.orgs
    ADD CONSTRAINT orgs_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_pkey PRIMARY KEY (id);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_email_key UNIQUE (email);
ALTER TABLE ONLY public.users
    ADD CONSTRAINT users_org_id_fkey FOREIGN KEY (org_id) REFERENCES public.orgs(id) ON DELETE CASCADE;
ALTER TABLE ONLY billing.invoices
    ADD CONSTRAINT invoices_user_fk FOREIGN KEY (user_id) REFERENCES public.users(id);

CREATE INDEX idx_users_org ON public.users USING btree (org_id, created_at);
CREATE UNIQUE INDEX idx_users_lower ON public.users USING btree (lower((email)::text));
`

const myDump = "-- MySQL dump 10.13\n" +
	"/*!40101 SET @saved_cs_client     = @@character_set_client */;\n" +
	"DROP TABLE IF EXISTS `posts`;\n" +
	"CREATE TABLE `posts` (\n" +
	"  `id` int unsigned NOT NULL AUTO_INCREMENT,\n" +
	"  `author_id` int unsigned NOT NULL COMMENT 'Who wrote it',\n" +
	"  `title` varchar(200) COLLATE utf8mb4_unicode_ci DEFAULT NULL,\n" +
	"  `state` enum('draft','live') NOT NULL DEFAULT 'draft',\n" +
	"  `body` text,\n" +
	"  `created_at` datetime NOT NULL DEFAULT CURRENT_TIMESTAMP,\n" +
	"  PRIMARY KEY (`id`),\n" +
	"  UNIQUE KEY `uq_title` (`title`),\n" +
	"  KEY `idx_author` (`author_id`),\n" +
	"  FULLTEXT KEY `ft_body` (`body`),\n" +
	"  CONSTRAINT `fk_author` FOREIGN KEY (`author_id`) REFERENCES `authors` (`id`) ON DELETE SET NULL ON UPDATE CASCADE\n" +
	") ENGINE=InnoDB AUTO_INCREMENT=5 DEFAULT CHARSET=utf8mb4 COMMENT='Blog posts';\n"

const liteDDL = `
CREATE TABLE IF NOT EXISTS "tags" (id INTEGER PRIMARY KEY AUTOINCREMENT, name TEXT UNIQUE NOT NULL);
CREATE TABLE post_tags (
  post_id INTEGER NOT NULL REFERENCES posts(id) ON DELETE CASCADE,
  tag_id INTEGER NOT NULL,
  PRIMARY KEY (post_id, tag_id),
  FOREIGN KEY (tag_id) REFERENCES tags (id)
) WITHOUT ROWID;
CREATE INDEX tag_idx ON post_tags(tag_id);
`

func table(t *testing.T, s *Schema, name string) *Table {
	t.Helper()
	for i := range s.Tables {
		if s.Tables[i].Name == name {
			return &s.Tables[i]
		}
	}
	t.Fatalf("no table %s", name)
	return nil
}

func col(t *testing.T, tb *Table, name string) *Column {
	t.Helper()
	for i := range tb.Columns {
		if tb.Columns[i].Name == name {
			return &tb.Columns[i]
		}
	}
	t.Fatalf("no column %s in %s", name, tb.Name)
	return nil
}

func TestPostgresDump(t *testing.T) {
	s, err := ParseDDL(pgDump)
	if err != nil {
		t.Fatal(err)
	}
	if len(s.Tables) != 3 {
		t.Fatalf("want 3 tables, got %d: %+v", len(s.Tables), s.Tables)
	}
	u := table(t, s, "users")
	if u.Comment != "Registered accounts" {
		t.Errorf("comment %q", u.Comment)
	}
	if strings.Join(u.PrimaryKey, ",") != "id" {
		t.Errorf("pk %v", u.PrimaryKey)
	}
	e := col(t, u, "email")
	if e.Type != "character varying(255)" || e.Nullable || !e.Unique || e.Comment != "Login e-mail, it's unique" {
		t.Errorf("email %+v", e)
	}
	b := col(t, u, "balance")
	if b.Type != "numeric(10,2)" || b.Default != "0.00" || b.Nullable {
		t.Errorf("balance %+v", b)
	}
	if m := col(t, u, "meta"); m.Default != `'{"a": "b,c"}'::jsonb` || !m.Nullable {
		t.Errorf("meta %+v", m)
	}
	if c := col(t, u, "created_at"); c.Type != "timestamp with time zone" || c.Default != "now()" {
		t.Errorf("created_at %+v", c)
	}
	if len(u.ForeignKeys) != 1 {
		t.Fatalf("fks %+v", u.ForeignKeys)
	}
	fk := u.ForeignKeys[0]
	if fk.RefTable != "orgs" || fk.Columns[0] != "org_id" || fk.RefColumns[0] != "id" || fk.OnDelete != "CASCADE" {
		t.Errorf("fk %+v", fk)
	}
	if len(u.Indexes) != 2 {
		t.Fatalf("indexes %+v", u.Indexes)
	}
	if strings.Join(u.Indexes[0].Columns, ",") != "org_id,created_at" || !u.Indexes[1].Unique {
		t.Errorf("indexes %+v", u.Indexes)
	}
	inv := table(t, s, "billing.invoices")
	if inv.ForeignKeys[0].RefTable != "users" {
		t.Errorf("schema prefix: %+v", inv.ForeignKeys[0])
	}
	if len(s.Enums) != 1 || s.Enums[0].Name != "order_status" || strings.Join(s.Enums[0].Values, ",") != "new,paid" {
		t.Errorf("enums %+v", s.Enums)
	}
	if len(s.Views) != 1 || s.Views[0] != "active_users" {
		t.Errorf("views %+v", s.Views)
	}
}

func TestMySQLDump(t *testing.T) {
	s, err := ParseDDL(myDump)
	if err != nil {
		t.Fatal(err)
	}
	p := table(t, s, "posts")
	if p.Comment != "Blog posts" {
		t.Errorf("comment %q", p.Comment)
	}
	if strings.Join(p.PrimaryKey, ",") != "id" {
		t.Errorf("pk %v", p.PrimaryKey)
	}
	if c := col(t, p, "id"); c.Type != "int unsigned" || c.Nullable || !c.AutoIncrement {
		t.Errorf("id %+v", c)
	}
	if c := col(t, p, "author_id"); c.Comment != "Who wrote it" {
		t.Errorf("author %+v", c)
	}
	if c := col(t, p, "title"); c.Type != "varchar(200)" || !c.Nullable || !c.Unique {
		t.Errorf("title %+v", c)
	}
	if c := col(t, p, "state"); c.Type != "enum('draft','live')" || c.Default != "'draft'" {
		t.Errorf("state %+v", c)
	}
	fk := p.ForeignKeys[0]
	if fk.RefTable != "authors" || fk.OnDelete != "SET NULL" || fk.OnUpdate != "CASCADE" {
		t.Errorf("fk %+v", fk)
	}
	if len(p.Indexes) != 2 {
		t.Errorf("want idx_author and ft_body, got %+v", p.Indexes)
	}
}

func TestSQLiteDDL(t *testing.T) {
	s, err := ParseDDL(liteDDL)
	if err != nil {
		t.Fatal(err)
	}
	tg := table(t, s, "tags")
	if strings.Join(tg.PrimaryKey, ",") != "id" || !col(t, tg, "id").AutoIncrement {
		t.Errorf("tags %+v", tg)
	}
	if c := col(t, tg, "name"); !c.Unique || c.Nullable {
		t.Errorf("name %+v", c)
	}
	pt := table(t, s, "post_tags")
	if strings.Join(pt.PrimaryKey, ",") != "post_id,tag_id" || len(pt.ForeignKeys) != 2 {
		t.Errorf("post_tags %+v", pt)
	}
	if len(pt.Indexes) != 1 {
		t.Errorf("idx %+v", pt.Indexes)
	}
}

func TestBadURLDoesNotLeakPassword(t *testing.T) {
	_, _, _, err := dumpCommand("postgres://u:hunter2@h:bad/db")
	if err == nil || strings.Contains(err.Error(), "hunter2") {
		t.Fatalf("%v", err)
	}
}

func TestColumnsNamedLikeKeywords(t *testing.T) {
	s, err := ParseDDL(`CREATE TABLE settings (
		key varchar(64) NOT NULL,
		index integer,
		check text,
		value text,
		unique_id bigint,
		KEY idx_value (value(20)),
		PRIMARY KEY (key)
	);`)
	if err != nil {
		t.Fatal(err)
	}
	tb := table(t, s, "settings")
	if len(tb.Columns) != 5 {
		t.Fatalf("columns %+v", tb.Columns)
	}
	if len(tb.Indexes) != 1 || strings.Join(tb.Indexes[0].Columns, ",") != "value" {
		t.Fatalf("indexes %+v", tb.Indexes)
	}
	if strings.Join(tb.PrimaryKey, ",") != "key" {
		t.Fatalf("pk %v", tb.PrimaryKey)
	}
}

func TestPostgresSerialAndArrays(t *testing.T) {
	s, err := ParseDDL(`
CREATE TABLE public.a (
    id integer NOT NULL,
    uid bigint NOT NULL,
    tags text[] DEFAULT '{}'::text[],
    scores numeric(4,1)[],
    total numeric GENERATED ALWAYS AS (1 + 1) STORED,
    kind public.order_status
);
ALTER TABLE ONLY public.a ALTER COLUMN id SET DEFAULT nextval('public.a_id_seq'::regclass);
ALTER TABLE public.a ALTER COLUMN uid ADD GENERATED BY DEFAULT AS IDENTITY (
    SEQUENCE NAME public.a_uid_seq
    START WITH 1
);
`)
	if err != nil {
		t.Fatal(err)
	}
	a := table(t, s, "a")
	if c := col(t, a, "id"); !c.AutoIncrement || c.Default != "" {
		t.Errorf("id %+v", c)
	}
	if c := col(t, a, "uid"); !c.AutoIncrement {
		t.Errorf("uid %+v", c)
	}
	if c := col(t, a, "tags"); c.Type != "text[]" || c.Default != "'{}'::text[]" {
		t.Errorf("tags %+v", c)
	}
	if c := col(t, a, "scores"); c.Type != "numeric(4,1)[]" {
		t.Errorf("scores %+v", c)
	}
	if c := col(t, a, "total"); c.AutoIncrement {
		t.Errorf("generated column is not an identity: %+v", c)
	}
	if c := col(t, a, "kind"); c.Type != "order_status" {
		t.Errorf("kind %+v", c)
	}
}

func TestPhpMyAdminStyleAlters(t *testing.T) {
	s, err := ParseDDL("CREATE TABLE `t` (`id` int NOT NULL, `a` int NOT NULL, `b` int DEFAULT NULL);\n" +
		"ALTER TABLE `t`\n  ADD PRIMARY KEY (`id`),\n  ADD UNIQUE KEY `uq` (`a`,`b`),\n  ADD KEY `k` (`b`);\n" +
		"ALTER TABLE `t`\n  MODIFY `id` int NOT NULL AUTO_INCREMENT;\n" +
		"ALTER TABLE `t`\n  ADD CONSTRAINT `fk` FOREIGN KEY (`b`) REFERENCES `o` (`id`) ON DELETE CASCADE;\n")
	if err != nil {
		t.Fatal(err)
	}
	tb := table(t, s, "t")
	if strings.Join(tb.PrimaryKey, ",") != "id" || len(tb.Indexes) != 2 || !tb.Indexes[0].Unique || tb.Indexes[1].Name != "k" {
		t.Errorf("%+v", tb)
	}
	if c := col(t, tb, "id"); !c.AutoIncrement || c.Nullable {
		t.Errorf("id %+v", c)
	}
	if len(tb.ForeignKeys) != 1 || tb.ForeignKeys[0].RefTable != "o" {
		t.Errorf("fk %+v", tb.ForeignKeys)
	}
}

func TestCorruptSQLiteNeverPanics(t *testing.T) {
	src, err := os.ReadFile(filepath.Join("testdata", "tiny_pages.db"))
	if err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(t.TempDir(), "c.db")
	x := uint32(2463534242)
	for range 500 {
		b := append([]byte(nil), src...)
		for range 8 {
			x ^= x << 13
			x ^= x >> 17
			x ^= x << 5
			b[int(x)%(len(b))] = byte(x >> 8)
		}
		b = append(b[:16:16], b[16:]...)
		os.WriteFile(p, b, 0o644)
		ReadSQLite(p)
	}
}

func TestMarkdownFull(t *testing.T) {
	s, _ := ParseDDL(pgDump)
	md := Markdown(s, MarkdownOptions{})
	for _, want := range []string{
		"## users\n",
		"Registered accounts\n",
		"- id bigint PK\n",
		"- org_id bigint -> orgs.id ON DELETE CASCADE\n",
		"- email character varying(255) NOT NULL UNIQUE -- Login e-mail, it's unique\n",
		"- balance numeric(10,2) NOT NULL DEFAULT 0.00\n",
		"Indexes: idx_users_org (org_id, created_at); UNIQUE idx_users_lower (lower((email)::text))\n",
		"## billing.invoices\n",
		"Enums:\n- order_status: new, paid\n",
		"Views: active_users\n",
	} {
		if !strings.Contains(md, want) {
			t.Errorf("missing %q in:\n%s", want, md)
		}
	}
}

func TestMarkdownCompact(t *testing.T) {
	s, _ := ParseDDL(pgDump)
	md := Markdown(s, MarkdownOptions{Compact: true})
	want := "users(id PK, org_id -> orgs.id, email, balance, meta, created_at)\n"
	if !strings.Contains(md, want) {
		t.Errorf("missing %q in:\n%s", want, md)
	}
	if strings.Contains(md, "bigint") {
		t.Errorf("compact must drop types:\n%s", md)
	}
}

func TestFilter(t *testing.T) {
	s, _ := ParseDDL(pgDump)
	if got := Markdown(s, MarkdownOptions{Only: []string{"users"}}); strings.Contains(got, "## orgs") || !strings.Contains(got, "## users") {
		t.Errorf("only:\n%s", got)
	}
	if got := Markdown(s, MarkdownOptions{Exclude: []string{"billing.*"}}); strings.Contains(got, "invoices") {
		t.Errorf("exclude:\n%s", got)
	}
}

func TestReadSQLite(t *testing.T) {
	for _, f := range []string{"small.db", "large.db", "tiny_pages.db"} {
		t.Run(f, func(t *testing.T) {
			s, err := ReadSQLite(filepath.Join("testdata", f))
			if err != nil {
				t.Fatal(err)
			}
			want := map[string]int{"small.db": 3, "large.db": 400, "tiny_pages.db": 40}[f]
			if len(s.Tables) != want {
				t.Fatalf("want %d tables, got %d", want, len(s.Tables))
			}
			tb := table(t, s, "wide")
			if len(tb.Columns) != 120 {
				t.Fatalf("overflow record truncated: %d columns", len(tb.Columns))
			}
			if len(table(t, s, "child").Indexes) != 1 {
				t.Fatal("index from sqlite_schema missing")
			}
		})
	}
}

func TestReadSQLiteRejectsOtherFiles(t *testing.T) {
	p := filepath.Join(t.TempDir(), "x.db")
	os.WriteFile(p, []byte("not a database"), 0o644)
	if _, err := ReadSQLite(p); err == nil {
		t.Fatal("want error")
	}
}

func TestDumpCommand(t *testing.T) {
	cases := []struct {
		target, bin string
		args        []string
	}{
		{"postgres://u:p@h:5432/db", "pg_dump", []string{"--schema-only", "--no-owner", "--no-privileges", "postgres://u@h:5432/db"}},
		{"postgresql://h/db", "pg_dump", []string{"--schema-only", "--no-owner", "--no-privileges", "postgresql://h/db"}},
		{"mysql://root:pw@db.local:3307/shop", "mysqldump", []string{"--no-data", "--skip-comments", "--single-transaction", "-h", "db.local", "-P", "3307", "-u", "root", "shop"}},
	}
	for _, c := range cases {
		bin, args, env, err := dumpCommand(c.target)
		if err != nil {
			t.Fatal(err)
		}
		if bin != c.bin || strings.Join(args, " ") != strings.Join(c.args, " ") {
			t.Errorf("%s: %s %v", c.target, bin, args)
		}
		if c.bin == "mysqldump" && (len(env) != 1 || env[0] != "MYSQL_PWD=pw") {
			t.Errorf("password must travel via env, got %v", env)
		}
		if strings.Contains(c.target, "u:p@") && (len(env) != 1 || env[0] != "PGPASSWORD=p") {
			t.Errorf("password must travel via env, got %v", env)
		}
	}
}

func TestLoadDispatch(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "dump.sql")
	os.WriteFile(p, []byte(liteDDL), 0o644)
	s, err := Load(p)
	if err != nil || len(s.Tables) != 2 {
		t.Fatalf("%v %+v", err, s)
	}
	if _, err := Load(filepath.Join("testdata", "small.db")); err != nil {
		t.Fatal(err)
	}
}
