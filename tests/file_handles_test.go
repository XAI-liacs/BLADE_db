package tests

import (
	filehandlers "XAI-liacs/BLADE_db/file_handlers"
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
		fmt.Println(problem)
		t.Fatal("Incompatible file was read.")
	}
}

func TestFileHandlersConversationLogReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/Fourier_Unequality/run-LLaMEAgemma4_latest-fourier_uncertainty_C4-0",
		Suffix: "conversationlog.jsonl",
	}

	cl := filehandlers.ConversationLog{}
	content, err := cl.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(content) != 20 {
		t.Fatalf("Not all rows imported, got total length %v, expected 20.", len(content))
	}
}

func TestFileHandlersConversationLogFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}
	cl := filehandlers.ConversationLog{}
	_, err := cl.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}

func TestFileHandlersSolutionLogReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/Fourier_Unequality/run-LLaMEAgemma4_latest-fourier_uncertainty_C4-0",
		Suffix: "log.jsonl",
	}

	sl := filehandlers.SolutionLog{}
	content, err := sl.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
	if len(content) != 10 {
		t.Fatalf("Not all rows imported, got total length %v, expected 10.", len(content))
	}
}

func TestFileHandlersSolutionLogFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}
	sl := filehandlers.SolutionLog{}
	_, err := sl.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}

func TestFileHandlersProgressReadsFileProperly(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/Fourier_Unequality",
		Suffix: "progress.json",
	}

	p := filehandlers.Progress{}
	_, err := p.Read(details)
	if err != nil {
		t.Fatal(err.Error())
	}
}

func TestFileHandlersProgressFailsWithIncompatibleFile(t *testing.T) {
	details := filehandlers.FileDetails{
		Root:   "./test_files/method",
		Suffix: "llamea.json",
	}
	p := filehandlers.Progress{}
	_, err := p.Read(details)
	if err == nil {
		t.Fatal("Incompatible file was read.")
	}
}
