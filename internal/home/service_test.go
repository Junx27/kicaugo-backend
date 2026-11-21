package home

import (
	"context"
	"errors"
	"testing"
)

func TestServiceGet(t *testing.T) {
	t.Run("berhasil", func(t *testing.T) {
		content, err := NewService(&fakeRepository{content: sampleContent()}).Get(context.Background())
		if err != nil || content.Summary.LiveCount != 14 {
			t.Errorf("Get() = %+v, %v", content, err)
		}
	})

	t.Run("gagal", func(t *testing.T) {
		if _, err := NewService(&fakeRepository{err: errors.New("db down")}).Get(context.Background()); err == nil {
			t.Error("Get() ingin error")
		}
	})
}
