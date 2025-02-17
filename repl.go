package main

import "strings"

func cleanInput(text string) []string {
	
	trimmed := strings.TrimSpace(text)
	lowercase := strings.ToLower(trimmed)
	slices := strings.Fields(lowercase)
	return slices
}