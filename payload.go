package main

import (
	"math/rand"
	"strconv"
	"time"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type payload struct {
	Id int `yaml:"id"`

	Int    int               `yaml:"int"`
	Float  float32           `yaml:"float"`
	Bool   bool              `yaml:"bool"`
	String string            `yaml:"string"`
	Slice  []string          `yaml:"slice"`
	Map    map[string]string `yaml:"map"`

	CreatedAt time.Time `yaml:"created_at"`
	UpdatedAt time.Time `yaml:"updated_at"`
}

func generateRandomPayload() payload {
	var (
		randomSliceSize = rand.Intn(5) + 1
		randomSlice     = make([]string, randomSliceSize)
	)

	for i := range randomSliceSize {
		randomSlice[i] = generateRandomString(rand.Intn(8) + 3)
	}

	var (
		randomMapSize = rand.Intn(5) + 1
		randomMap     = make(map[string]string)
	)

	for i := range randomMapSize {
		randomMap[strconv.Itoa(i)] = generateRandomString(rand.Intn(8) + 3)
	}

	return payload{
		Id:        rand.Intn(1000),
		Int:       rand.Intn(1000),
		Float:     rand.Float32() * 100,
		Bool:      rand.Intn(2) == 1,
		String:    generateRandomString(10),
		Slice:     randomSlice,
		Map:       randomMap,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func generateRandomString(size int) string {
	data := make([]rune, size)

	for i := range data {
		data[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(data)
}
