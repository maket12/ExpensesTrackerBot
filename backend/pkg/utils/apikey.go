package utils

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
)

// GenerateAPIKey возвращает:
//  1. сырой API-ключ, который мы выдаем ресторану один раз (32 байта случайных данных, base64 url без паддинга);
//  2. SHA-256 хеш, который сохраняется в БД в колонке api_key_hash.
func GenerateAPIKey() (raw string, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}

	raw = base64.RawURLEncoding.EncodeToString(buf)
	hash = HashToken(raw)
	return raw, hash, nil
}

// CompareAPIKeyHash сравнивает переданный сырой API-ключ с сохранённым в БД
// хешем в constant-time, чтобы избежать утечки через тайминг-атаки.
//
// Сравнение идёт через хеш, чтобы длина обеих строк была одинаковой
// (hex-кодировка SHA-256 — всегда 64 символа), иначе ConstantTimeCompare
// сам по себе зависит от длины.
//
// Возвращает true, только если raw совпадает с хешем, лежащим в expectedHash.
func CompareAPIKeyHash(raw, expectedHash string) bool {
	actualHash := HashToken(raw)
	return subtle.ConstantTimeCompare([]byte(actualHash), []byte(expectedHash)) == 1
}
