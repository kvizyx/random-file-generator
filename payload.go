package main

import (
	"math/rand"
	"strconv"
)

var alphabet = []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

type payload struct {
	Int    int               `yaml:"int"`
	Float  float32           `yaml:"float"`
	Bool   bool              `yaml:"bool"`
	String string            `yaml:"string"`
	Slice  []string          `yaml:"slice"`
	Map    map[string]string `yaml:"map"`
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
		Int:    rand.Intn(1000),
		Float:  rand.Float32() * 100,
		Bool:   rand.Intn(2) == 1,
		String: generateRandomString(10),
		Slice:  randomSlice,
		Map:    randomMap,
	}
}

func generateRandomString(size int) string {
	data := make([]rune, size)

	for i := range data {
		data[i] = alphabet[rand.Intn(len(alphabet))]
	}

	return string(data)
}
