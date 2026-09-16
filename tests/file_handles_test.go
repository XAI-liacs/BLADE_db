package tests

import (
	filehandlers "anantashahane/BLADE_db/file_handlers"
	"fmt"
	"testing"
)

func TestFileHandlersLLMReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/llm",
		Suffix: "gemma_1.json",
	}

	llm := filehandlers.LLM{}
	content, err := llm.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
	if content[0].Model != "gemma4:latest" {
		t.Fatal("Model name miss-match.")
	}
	if !content[0].Hardware.Valid {
		t.Fatal("Invalid hardware validity assignment.")
	}
}

func TestFileHandlersLLMFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}
	llm := filehandlers.LLM{}
	_, err := llm.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}
func TestFileHandlersMethodReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}

	method := filehandlers.Method{}
	content, err := method.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
	if content.Name != "LLaMEA-gpt-oss:20b" {
		t.Fatal("Method name miss-match.")
	}
	if content.Source != "https://github.com/XAI-liacs/LLaMEA" {
		t.Fatal("Method source not right...")
	}
	fmt.Println(content)
}

func TestFileHandlersMethodFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/llm",
		Suffix: "gemma_1.json",
	}
	method := filehandlers.Method{}
	_, err := method.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}
func TestFileHandlersProblemReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/problems",
		Suffix: "Auto_corr_1_problem.json",
	}

	problem := filehandlers.Problem{}
	content, err := problem.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(content.Tags) != 4 {
		t.Fatal("Tags not properly imported.")
	}
	if content.Problem.Name != "Auto-Correlation 1" {
		t.Fatal("Name import fail.")
	}
}

func TestFileHandlersProblemFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}
	problem := filehandlers.Problem{}
	_, err := problem.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}
