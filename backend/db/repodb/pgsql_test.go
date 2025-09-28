package repodb

import (
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

func getTestDBConnAttrs() DBConnAttrs {
	return DBConnAttrs{
		port:     os.Getenv("TEST_DB_PORT"),
		user:     os.Getenv("TEST_DB_USER"),
		password: os.Getenv("TEST_DB_PASSWORD"),
		dbname:   os.Getenv("TEST_DB_NAME"),
		sslmode:  os.Getenv("TEST_DB_SSLMODE"),
	}
}

func setupTestDB(t *testing.T) *PGSQLRepo {
	t.Helper()

	repo, err := NewPGSQLRepo(getTestDBConnAttrs())
	if err != nil {
		t.Fatalf("Failed to create repo: %v", err)
	}

	_, err = repo.db.Exec("DELETE FROM documents")
	if err != nil {
		t.Fatalf("Failed to clean table: %v", err)
	}

	return repo
}

func TestMain(m *testing.M) {
	err := godotenv.Load("../../.env")
	if err != nil {
		panic(fmt.Sprintf("Can not load evironment variables: %v", err))
	}

	code := m.Run()
	os.Exit(code)
}

func TestNewPGSQLRepo(t *testing.T) {
	tests := []struct {
		name    string
		attrs   DBConnAttrs
		wantErr bool
	}{
		{
			name:    "valid connection",
			attrs:   getTestDBConnAttrs(),
			wantErr: false,
		},
		{
			name: "invalid connection",
			attrs: DBConnAttrs{
				user:     "invalid",
				password: "invalid",
				dbname:   "invalid",
				sslmode:  "disable",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := NewPGSQLRepo(tt.attrs)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewPGSQLRepo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && repo == nil {
				t.Error("NewPGSQLRepo() returned nil repo without error")
			}

			if repo != nil {
				repo.Close()
			}
		})
	}
}

func TestPGSQLRepo_CreateAndGet(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "test.txt"
	data := []byte("test data")

	err := repo.Create(title, data)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	retrievedData, err := repo.Get(title)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if string(retrievedData) != string(data) {
		t.Errorf("Get() = %s, want %s", string(retrievedData), string(data))
	}
}

func TestPGSQLRepo_CreateDuplicate(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "test.txt"
	data := []byte("test data")

	err := repo.Create(title, data)
	if err != nil {
		t.Fatalf("First Create failed: %v", err)
	}

	err = repo.Create(title, data)
	if err != ErrFileExists {
		t.Errorf("Create duplicate should return ErrFileExists, got: %v", err)
	}
}

func TestPGSQLRepo_GetNonExistent(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	_, err := repo.Get("nonexistent.txt")
	if err != ErrFileNotFound {
		t.Errorf("Get non-existent should return ErrFileNotFound, got: %v", err)
	}
}

func TestPGSQLRepo_Save(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "test.txt"
	initialData := []byte("initial data")
	updatedData := []byte("updated data")

	err := repo.Create(title, initialData)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Save(title, updatedData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrievedData, err := repo.Get(title)
	if err != nil {
		t.Fatalf("Get after Save failed: %v", err)
	}

	if string(retrievedData) != string(updatedData) {
		t.Errorf("Save() didn't update data, got %s, want %s",
			string(retrievedData), string(updatedData))
	}
}

func TestPGSQLRepo_SaveNotExistent(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "nonexistent.txt"
	data := []byte("nonexistent data")

	err := repo.Save(title, data)
	if err != ErrFileNotFound {
		t.Errorf("Save non-existent should return ErrFileNotFound, got: %v", err)
	}
}

func TestPGSQLRepo_Delete(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "test.txt"
	data := []byte("test data")

	err := repo.Create(title, data)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	err = repo.Delete(title)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get(title)
	if err != ErrFileNotFound {
		t.Errorf("Get after Delete should return ErrFileNotFound, got: %v", err)
	}
}

func TestPGSQLRepo_DeleteNonExistent(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	err := repo.Delete("nonexistent.txt")
	if err != ErrFileNotFound {
		t.Errorf("Delete non-existent should return ErrFileNotFound, got: %v", err)
	}
}

func TestPGSQLRepo_GetList(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	files := map[string][]byte{
		"file1.txt": []byte("data1"),
		"file2.txt": []byte("data2"),
		"file3.txt": []byte("data3"),
	}

	for title, data := range files {
		err := repo.Create(title, data)
		if err != nil {
			t.Fatalf("Create %s failed: %v", title, err)
		}
	}

	titles, err := repo.GetList()
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if len(titles) != len(files) {
		t.Errorf("GetList() returned %d files, want %d", len(titles), len(files))
	}

	titleMap := make(map[string]bool)
	for _, title := range titles {
		titleMap[title] = true
	}

	for expectedTitle := range files {
		if !titleMap[expectedTitle] {
			t.Errorf("GetList() missing file: %s", expectedTitle)
		}
	}
}

func TestPGSQLRepo_GetListEmpty(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	titles, err := repo.GetList()
	if err != nil {
		t.Fatalf("GetList on empty table failed: %v", err)
	}

	if len(titles) != 0 {
		t.Errorf("GetList() on empty table returned %d files, want 0", len(titles))
	}
}

func TestPGSQLRepo_ConcurrentAccess(t *testing.T) {
	repo := setupTestDB(t)
	defer repo.Close()

	title := "concurrent.txt"
	data := []byte("test data")

	err := repo.Create(title, data)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	done := make(chan bool, 5)
	for i := 0; i < 5; i++ {
		go func() {
			_, err := repo.Get(title)
			if err != nil {
				t.Errorf("Concurrent Get failed: %v", err)
			}
			done <- true
		}()
	}

	for i := 0; i < 5; i++ {
		<-done
	}
}

func TestPGSQLRepo_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	repo := setupTestDB(t)
	defer repo.Close()

	testData := []struct {
		title string
		data  []byte
	}{
		{"doc1.txt", []byte("content1")},
		{"doc2.txt", []byte("content2")},
		{"doc3.txt", []byte("content3")},
	}

	for _, td := range testData {
		err := repo.Create(td.title, td.data)
		if err != nil {
			t.Fatalf("Create %s failed: %v", td.title, err)
		}
	}

	titles, err := repo.GetList()
	if err != nil {
		t.Fatalf("GetList failed: %v", err)
	}

	if len(titles) != len(testData) {
		t.Errorf("Expected %d files, got %d", len(testData), len(titles))
	}

	updatedData := []byte("updated content")
	err = repo.Save("doc2.txt", updatedData)
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	retrieved, err := repo.Get("doc2.txt")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}

	if string(retrieved) != string(updatedData) {
		t.Errorf("Update failed, got %s, want %s", string(retrieved), string(updatedData))
	}

	err = repo.Delete("doc1.txt")
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	_, err = repo.Get("doc1.txt")
	if err != ErrFileNotFound {
		t.Errorf("Expected ErrFileNotFound after deletion, got: %v", err)
	}

	finalTitles, err := repo.GetList()
	if err != nil {
		t.Fatalf("Final GetList failed: %v", err)
	}

	if len(finalTitles) != len(testData)-1 {
		t.Errorf("Expected %d files after deletion, got %d", len(testData)-1, len(finalTitles))
	}
}
