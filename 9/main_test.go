package main

import (
	"errors"
	"testing"
)

func TestUnpackStringCorrect(t *testing.T) {
	expected := []string{"aaaabccddddde", "abcd", "", "qwe45", "qwe44444", "44bb"}
	input := []string{"a4bc2d5e", "abcd", "", "qwe\\4\\5", "qwe\\45", "\\42b2"}
	for i, v := range input {
		res, err := UnpackString(v)
		if err != nil {
			t.Errorf("Ошибка: %v\n", err)
		}
		if res != expected[i] {
			t.Errorf("Ввод: %s, Ожидалось: %s, Получилось: %s", v, expected[i], res)
		}
	}

}

func TestUnpackStringIncorrect(t *testing.T) {
	input := []string{"45", "\\33bb3245"}
	for _, v := range input {
		_, err := UnpackString(v)
		if errors.Is(err, nil) {
			t.Error("Ожидалась ошибка")
		}
	}
}
