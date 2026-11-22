// Command scaffold generates a CRUD feature skeleton for KicauGo.
package main

import (
	"embed"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

type field struct {
	GoName  string
	GoType  string
	SQLType string
	JSONTag string
}

type data struct {
	Domain       string
	DomainPascal string
	Table        string
	Fields       []field
	FieldColumns string
	Ops          map[string]bool
}

var typeMap = map[string]struct {
	goType  string
	sqlType string
}{
	"string":  {"string", "TEXT"},
	"int":     {"int", "INTEGER"},
	"int64":   {"int64", "BIGINT"},
	"bool":    {"bool", "BOOLEAN"},
	"float64": {"float64", "DOUBLE PRECISION"},
	"time":    {"time.Time", "TIMESTAMPTZ"},
}

func main() {
	name := flag.String("name", "", "domain name (lowercase, single word), e.g. product")
	fields := flag.String("fields", "name:string", "comma-separated fields, e.g. name:string,price:int,active:bool")
	table := flag.String("table", "", "table name (default: domain name)")
	opsFlag := flag.String("ops", "crud", "operations: crud/read/write or comma list of create,list,get,update,delete")
	flag.Parse()

	if *name == "" {
		fail("flag -name wajib diisi")
	}
	if !validName(*name) {
		fail("nama domain harus huruf kecil a-z tanpa spasi/underscore, contoh: product")
	}

	domain := *name
	domainPascal := capitalize(domain)
	tbl := *table
	if tbl == "" {
		tbl = domain
	}

	var fieldList []field
	if strings.TrimSpace(*fields) != "" {
		fl, err := parseFields(*fields)
		if err != nil {
			fail(err.Error())
		}
		fieldList = fl
	}

	ops, err := parseOps(*opsFlag)
	if err != nil {
		fail(err.Error())
	}

	cols := make([]string, len(fieldList))
	for i, f := range fieldList {
		cols[i] = f.JSONTag
	}
	d := data{
		Domain:       domain,
		DomainPascal: domainPascal,
		Table:        tbl,
		Fields:       fieldList,
		FieldColumns: strings.Join(cols, ", "),
		Ops:          ops,
	}

	featureDir := filepath.Join("internal", "features", domain)
	augment := false
	if _, err := os.Stat(featureDir); err == nil {
		augment = true
	}

	if augment {
		// Augment mode: tambahkan operasi ke fitur yang sudah ada.
		if !ops["delete"] {
			fail("augment mode saat ini mendukung OPS=delete. OPS lain belum didukung untuk augment.")
		}
		if err := augmentDelete(domain, domainPascal, d); err != nil {
			fail(err.Error())
		}
		files := []string{
			filepath.Join("internal", "features", domain, "repository.go"),
			filepath.Join("internal", "features", domain, "service.go"),
			filepath.Join("internal", "features", domain, "handler.go"),
		}
		if err := runGofmt(files); err != nil {
			fail("gofmt gagal: " + err.Error())
		}
		if err := runBuild(); err != nil {
			fail("build gagal setelah augment: " + err.Error())
		}
		fmt.Printf("Augment selesai: operasi delete ditambahkan ke '%s'.\n", domain)
		fmt.Println("Catatan: tambahkan test secara manual untuk operasi baru (repository/service/handler test).")
		return
	}

	var created []string
	rollback := func() {
		for _, f := range created {
			os.Remove(f)
		}
		os.Remove(featureDir)
	}

	// 1. SQL migration
	migPath, err := createMigration(domain, tbl, fieldList)
	if err != nil {
		fail(err.Error())
	}
	created = append(created, migPath)

	// 2. entity
	entityPath := filepath.Join("internal", "entities", domain+".go")
	if err := renderTemplate("entity.go.tmpl", entityPath, d); err != nil {
		rollback()
		fail(err.Error())
	}
	created = append(created, entityPath)

	// 3. feature dir + files
	if err := os.MkdirAll(featureDir, 0755); err != nil {
		rollback()
		fail(err.Error())
	}
	featureFiles := []string{"repository.go", "repository_test.go", "service.go", "service_test.go", "handler.go", "handler_test.go"}
	for _, fn := range featureFiles {
		outPath := filepath.Join(featureDir, fn)
		if err := renderTemplate(fn+".tmpl", outPath, d); err != nil {
			rollback()
			fail(err.Error())
		}
		created = append(created, outPath)
	}

	// 3b. request/response DTOs in utils
	reqDir := filepath.Join("internal", "utils", "request")
	respDir := filepath.Join("internal", "utils", "response")
	reqPath := filepath.Join(reqDir, domain+".go")
	respPath := filepath.Join(respDir, domain+".go")
	if _, err := os.Stat(reqPath); err == nil {
		rollback()
		fail("file sudah ada: " + reqPath)
	}
	if _, err := os.Stat(respPath); err == nil {
		rollback()
		fail("file sudah ada: " + respPath)
	}
	if err := renderTemplate("request.go.tmpl", reqPath, d); err != nil {
		rollback()
		fail(err.Error())
	}
	created = append(created, reqPath)
	if err := renderTemplate("response.go.tmpl", respPath, d); err != nil {
		rollback()
		fail(err.Error())
	}
	created = append(created, respPath)

	// 4. wiring injection (backup first)
	backupMain, _ := os.ReadFile("cmd/api/main.go")
	backupServer, _ := os.ReadFile("cmd/server/server.go")
	if err := injectMain(domain, domainPascal); err != nil {
		rollback()
		fail("gagal inject main.go: " + err.Error())
	}
	if err := injectServer(domain, domainPascal); err != nil {
		rollback()
		os.WriteFile("cmd/api/main.go", backupMain, 0644)
		fail("gagal inject server.go: " + err.Error())
	}

	// 5. gofmt
	if err := runGofmt(created); err != nil {
		rollback()
		os.WriteFile("cmd/api/main.go", backupMain, 0644)
		os.WriteFile("cmd/server/server.go", backupServer, 0644)
		fail("gofmt gagal: " + err.Error())
	}

	// 6. build check
	if err := runBuild(); err != nil {
		rollback()
		os.WriteFile("cmd/api/main.go", backupMain, 0644)
		os.WriteFile("cmd/server/server.go", backupServer, 0644)
		fail("build gagal setelah scaffold: " + err.Error())
	}

	fmt.Printf("Scaffold selesai untuk domain '%s'.\n", domain)
	fmt.Println("File yang dibuat:")
	for _, f := range created {
		fmt.Println("  " + f)
	}
	fmt.Println("Wiring diinjeksi ke cmd/api/main.go dan cmd/server/server.go.")
	fmt.Println("Jalankan: go test -race ./internal/features/" + domain)
}

func parseOps(s string) (map[string]bool, error) {
	ops := map[string]bool{}
	s = strings.TrimSpace(s)
	if s == "" {
		s = "crud"
	}
	// shorthand
	switch s {
	case "crud":
		for _, o := range []string{"create", "list", "get", "update", "delete"} {
			ops[o] = true
		}
		return ops, nil
	case "read":
		ops["list"] = true
		ops["get"] = true
		return ops, nil
	case "write":
		ops["create"] = true
		ops["update"] = true
		ops["delete"] = true
		return ops, nil
	}
	for _, part := range strings.Split(s, ",") {
		part = strings.TrimSpace(strings.ToLower(part))
		if part == "" {
			continue
		}
		switch part {
		case "create", "list", "get", "update", "delete":
			ops[part] = true
		default:
			return nil, fmt.Errorf("operasi tidak dikenal: %s", part)
		}
	}
	if len(ops) == 0 {
		return nil, fmt.Errorf("minimal satu operasi diperlukan")
	}
	return ops, nil
}
func validName(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !(r >= 'a' && r <= 'z') {
			return false
		}
	}
	return true
}

func parseFields(s string) ([]field, error) {
	parts := strings.Split(s, ",")
	var out []field
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p == "" {
			continue
		}
		kv := strings.Split(p, ":")
		if len(kv) != 2 {
			return nil, fmt.Errorf("format field salah: %s (harus name:type)", p)
		}
		name := strings.TrimSpace(kv[0])
		typ := strings.TrimSpace(kv[1])
		if name == "" {
			return nil, fmt.Errorf("nama field kosong")
		}
		if !validFieldName(name) {
			return nil, fmt.Errorf("nama field harus PascalCase: %s", name)
		}
		m, ok := typeMap[typ]
		if !ok {
			return nil, fmt.Errorf("tipe tidak dikenal: %s (dukung: string, int, int64, bool, float64, time)", typ)
		}
		out = append(out, field{
			GoName:  name,
			GoType:  m.goType,
			SQLType: m.sqlType,
			JSONTag: toSnake(name),
		})
	}
	if len(out) == 0 {
		return nil, fmt.Errorf("minimal satu field diperlukan")
	}
	return out, nil
}

func validFieldName(s string) bool {
	if s == "" {
		return false
	}
	if !(s[0] >= 'A' && s[0] <= 'Z') {
		return false
	}
	for _, r := range s {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9')) {
			return false
		}
	}
	return true
}

func capitalize(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

func toSnake(s string) string {
	var b strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			b.WriteByte('_')
		}
		b.WriteRune(r)
	}
	return strings.ToLower(b.String())
}

func createMigration(domain, table string, fields []field) (string, error) {
	last, _ := lastNumber("internal/database/migrations", "*.sql")
	next := last + 1
	fileName := fmt.Sprintf("%04d_%s.sql", next, domain+"_table")
	path := filepath.Join("internal", "database", "migrations", fileName)

	tmpl, err := template.ParseFS(templateFS, "templates/migration.sql.tmpl")
	if err != nil {
		return "", err
	}
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	d := data{Domain: domain, DomainPascal: capitalize(domain), Table: table, Fields: fields}
	if err := tmpl.Execute(f, d); err != nil {
		return "", err
	}
	return path, nil
}

func lastNumber(dir, pattern string) (int, error) {
	entries, err := filepath.Glob(filepath.Join(dir, pattern))
	if err != nil {
		return 0, err
	}
	sort.Strings(entries)
	if len(entries) == 0 {
		return 0, nil
	}
	last := entries[len(entries)-1]
	base := filepath.Base(last)
	num := 0
	for _, r := range base {
		if r >= '0' && r <= '9' {
			num = num*10 + int(r-'0')
		} else {
			break
		}
	}
	return num, nil
}

func renderTemplate(tmplName, outPath string, d data) error {
	tmpl, err := template.New(tmplName).Funcs(template.FuncMap{
		"add": func(a, b int) int { return a + b },
	}).ParseFS(templateFS, "templates/"+tmplName)
	if err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return tmpl.Execute(f, d)
}

func injectMain(domain, domainPascal string) error {
	path := "cmd/api/main.go"
	repl := map[string]string{
		"// scaffold:imports":  fmt.Sprintf("\t\"github.com/Junx27/kicaugo-backend/internal/features/%s\"\n// scaffold:imports", domain),
		"// scaffold:services": fmt.Sprintf("\t%sService := %s.NewService(%s.NewRepository(pool))\n// scaffold:services", domainPascal, domain, domain),
		"// scaffold:deps":     fmt.Sprintf("\t\t%s: %sService,\n// scaffold:deps", domainPascal, domainPascal),
	}
	return injectFile(path, repl)
}

func injectServer(domain, domainPascal string) error {
	path := "cmd/server/server.go"
	repl := map[string]string{
		"// scaffold:imports": fmt.Sprintf("\t\"github.com/Junx27/kicaugo-backend/internal/features/%s\"\n// scaffold:imports", domain),
		"// scaffold:deps":    fmt.Sprintf("\t%s %s.Service\n// scaffold:deps", domainPascal, domain),
		"// scaffold:routes":  fmt.Sprintf("\t%s.RegisterRoutes(api, deps.%s)\n// scaffold:routes", domain, domainPascal),
	}
	return injectFile(path, repl)
}

func injectFile(path string, replacements map[string]string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(b)
	for marker, repl := range replacements {
		if !strings.Contains(content, marker) {
			return fmt.Errorf("marker tidak ditemukan: %s", marker)
		}
		content = strings.Replace(content, marker, repl, 1)
	}
	return os.WriteFile(path, []byte(content), 0644)
}

func runGofmt(files []string) error {
	var goFiles []string
	for _, f := range files {
		if strings.HasSuffix(f, ".go") {
			goFiles = append(goFiles, f)
		}
	}
	if len(goFiles) == 0 {
		return nil
	}
	args := append([]string{"-w"}, goFiles...)
	cmd := exec.Command("gofmt", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func runBuild() error {
	cmd := exec.Command("go", "build", "./...")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func augmentDelete(domain, domainPascal string, d data) error {
	repoPath := filepath.Join("internal", "features", domain, "repository.go")
	servicePath := filepath.Join("internal", "features", domain, "service.go")
	handlerPath := filepath.Join("internal", "features", domain, "handler.go")
	respPath := filepath.Join("internal", "utils", "response", domain+".go")

	// Pastikan import yang dibutuhkan ada sebelum menyisipkan kode yang memakainya.
	if b, err := os.ReadFile(repoPath); err == nil {
		c := string(b)
		if !strings.Contains(c, "utils/errors") {
			lines := strings.Split(c, "\n")
			for i, l := range lines {
				if strings.Contains(l, "internal/entities") {
					lines[i] = l + "\n\tapperrors \"github.com/Junx27/kicaugo-backend/internal/utils/errors\""
					break
				}
			}
			os.WriteFile(repoPath, []byte(strings.Join(lines, "\n")), 0644)
		}
	}
	if b, err := os.ReadFile(handlerPath); err == nil {
		c := string(b)
		if !strings.Contains(c, "\"errors\"") {
			c = strings.Replace(c, "import (\n\t\"context\"\n\t\"net/http\"", "import (\n\t\"context\"\n\t\"errors\"\n\t\"net/http\"", 1)
		}
		if !strings.Contains(c, "utils/errors") {
			lines := strings.Split(c, "\n")
			for i, l := range lines {
				if strings.Contains(l, "internal/utils/http") {
					lines[i] = l + "\n\tapperrors \"github.com/Junx27/kicaugo-backend/internal/utils/errors\""
					break
				}
			}
			c = strings.Join(lines, "\n")
		}
		if !strings.Contains(c, "utils/response") {
			lines := strings.Split(c, "\n")
			for i, l := range lines {
				if strings.Contains(l, "internal/utils/http") {
					lines[i] = l + "\n\t\"github.com/Junx27/kicaugo-backend/internal/utils/response\""
					break
				}
			}
			c = strings.Join(lines, "\n")
		}
		os.WriteFile(handlerPath, []byte(c), 0644)
	}

	// Repository interface + impl (idempotent)
	if !strings.Contains(mustRead(repoPath), "func (r *repository) Delete") {
		repoIfaceSnippet := "\tDelete(ctx context.Context, id int64) error\n"
		repoSnippet := `
// Delete menghapus {{.Domain}}.
func (r *repository) Delete(ctx context.Context, id int64) error {
	const query = ` + "`DELETE FROM {{.Table}} WHERE id = $1`" + `
	tag, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("delete {{.Domain}} %d: %w", id, err)
	}
	if tag.RowsAffected() == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
`
		if err := insertBeforeMarker(repoPath, "// scaffold:repo-interface", repoIfaceSnippet, domain, domainPascal); err != nil {
			return err
		}
		if err := insertBeforeMarker(repoPath, "// scaffold:repo-methods", repoSnippet, domain, domainPascal); err != nil {
			return err
		}
	}

	// Service interface + impl (idempotent)
	if !strings.Contains(mustRead(servicePath), "func (s *service) Delete") {
		serviceIfaceSnippet := "\tDelete(ctx context.Context, id int64) error\n"
		serviceImplSnippet := `
// Delete menghapus {{.Domain}}.
func (s *service) Delete(ctx context.Context, id int64) error {
	return s.repo.Delete(ctx, id)
}
`
		if err := insertBeforeMarker(servicePath, "// scaffold:service-interface", serviceIfaceSnippet, domain, domainPascal); err != nil {
			return err
		}
		if err := insertBeforeMarker(servicePath, "// scaffold:service-methods", serviceImplSnippet, domain, domainPascal); err != nil {
			return err
		}
	}

	// Response DTO (idempotent)
	respPath2 := respPath
	if _, err := os.Stat(respPath2); err == nil {
		if !strings.Contains(mustRead(respPath2), "DeleteOutput") {
			deleteOutputStruct := `
// {{.DomainPascal}}DeleteOutput adalah output untuk DELETE.
type {{.DomainPascal}}DeleteOutput struct {
	Status int ` + "`" + `status:"204"` + "`" + `
}

`
			if err := insertBeforeMarker(respPath2, "// scaffold:response-types", deleteOutputStruct, domain, domainPascal); err != nil {
				return err
			}
		}
	}

	// Handler route (idempotent)
	handlerPath2 := handlerPath
	if !strings.Contains(mustRead(handlerPath2), "http.MethodDelete") {
		deleteRoute := `
	// DELETE /api/v1/{{.Domain}}s/{id}
	apphttp.Register(api, apphttp.Operation{
		ID:      "delete{{.DomainPascal}}",
		Method:  http.MethodDelete,
		Path:    "/api/v1/{{.Domain}}s/{id}",
		Summary: "Delete {{.DomainPascal}}",
		Tags:    []string{"{{.Domain}}"},
	}, func(ctx context.Context, input *struct {
		ID int64 ` + "`" + `path:"id"` + "`" + `
	}) (*response.{{.DomainPascal}}DeleteOutput, error) {
		if err := service.Delete(ctx, input.ID); err != nil {
			return nil, mapServiceError(err)
		}
		resp := &response.{{.DomainPascal}}DeleteOutput{}
		resp.Status = http.StatusNoContent
		return resp, nil
	})

`
		if err := insertBeforeMarker(handlerPath2, "// scaffold:routes", deleteRoute, domain, domainPascal); err != nil {
			return err
		}
	}

	// update mapServiceError to handle ErrNotFound
	b, err := os.ReadFile(handlerPath2)
	if err == nil {
		c := string(b)
		oldMap := `func mapServiceError(err error) error {
	return apphttp.InternalServerError("An error occurred. Please try again.")
}`
		newMap := `func mapServiceError(err error) error {
	if errors.Is(err, apperrors.ErrNotFound) {
		return apphttp.NotFound("{{.DomainPascal}} not found")
	}
	return apphttp.InternalServerError("An error occurred. Please try again.")
}`
		// newMap tidak melewati insertBeforeMarker, jadi placeholder harus dirender manual.
		newMap = strings.ReplaceAll(newMap, "{{.DomainPascal}}", domainPascal)
		if strings.Contains(c, oldMap) {
			c = strings.Replace(c, oldMap, newMap, 1)
		}
		if !strings.Contains(c, "\"errors\"") {
			c = strings.Replace(c, "import (\n\t\"context\"\n\t\"net/http\"", "import (\n\t\"context\"\n\t\"errors\"\n\t\"net/http\"", 1)
		}
		if !strings.Contains(c, "apperrors \"github.com/Junx27/kicaugo-backend/internal/utils/errors\"") {
			c = strings.Replace(c, "\tapphttp \"github.com/Junx27/kicaugo-backend/internal/utils/http\"\n", "\tapphttp \"github.com/Junx27/kicaugo-backend/internal/utils/http\"\n\tapperrors \"github.com/Junx27/kicaugo-backend/internal/utils/errors\"\n", 1)
		}
		if !strings.Contains(c, "github.com/Junx27/kicaugo-backend/internal/utils/response") {
			c = strings.Replace(c, "\tapphttp \"github.com/Junx27/kicaugo-backend/internal/utils/http\"\n", "\tapphttp \"github.com/Junx27/kicaugo-backend/internal/utils/http\"\n\t\"github.com/Junx27/kicaugo-backend/internal/utils/response\"\n", 1)
		}
		os.WriteFile(handlerPath2, []byte(c), 0644)
	}

	// apperrors import in repository.go
	repoBytes, _ := os.ReadFile(repoPath)
	repoContent := string(repoBytes)
	if !strings.Contains(repoContent, "apperrors") {
		repoContent = strings.Replace(repoContent, "\t\"github.com/Junx27/kicaugo-backend/internal/entities\"\n", "\t\"github.com/Junx27/kicaugo-backend/internal/entities\"\n\tapperrors \"github.com/Junx27/kicaugo-backend/internal/utils/errors\"\n", 1)
		os.WriteFile(repoPath, []byte(repoContent), 0644)
	}

	// Pastikan mock test mengimplementasikan Delete (agar go vet/test build).
	testFiles := []struct {
		path   string
		needle string
		method string
	}{
		{filepath.Join("internal", "features", domain, "repository_test.go"), "func (m *mockRepository) Delete", "func (m *mockRepository) Delete(ctx context.Context, id int64) error {\n\treturn nil\n}\n"},
		{filepath.Join("internal", "features", domain, "service_test.go"), "func (f *fakeRepository) Delete", "func (f *fakeRepository) Delete(ctx context.Context, id int64) error {\n\treturn nil\n}\n"},
		{filepath.Join("internal", "features", domain, "handler_test.go"), "func (f *fakeService) Delete", "func (f *fakeService) Delete(ctx context.Context, id int64) error {\n\treturn nil\n}\n"},
	}
	for _, tf := range testFiles {
		if b, err := os.ReadFile(tf.path); err == nil {
			c := string(b)
			if !strings.Contains(c, tf.needle) {
				c = c + "\n" + tf.method
				os.WriteFile(tf.path, []byte(c), 0644)
			}
		}
	}

	return nil
}
func mustRead(path string) string {
	b, _ := os.ReadFile(path)
	return string(b)
}

func insertBeforeMarker(path, marker, snippet, domain, domainPascal string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	content := string(b)
	if !strings.Contains(content, marker) {
		return fmt.Errorf("marker tidak ditemukan: %s di %s", marker, path)
	}
	rendered := snippet
	rendered = strings.ReplaceAll(rendered, "{{.Domain}}", domain)
	rendered = strings.ReplaceAll(rendered, "{{.DomainPascal}}", domainPascal)
	rendered = strings.ReplaceAll(rendered, "{{.Table}}", domain)
	// fields not needed for delete
	content = strings.Replace(content, marker, rendered+"\n"+marker, 1)
	return os.WriteFile(path, []byte(content), 0644)
}

func fail(msg string) {
	fmt.Fprintln(os.Stderr, "ERROR:", msg)
	os.Exit(1)
}
