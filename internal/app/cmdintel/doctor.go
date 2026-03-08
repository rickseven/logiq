package cmdintel

import (
	"os"
	"os/exec"

	"github.com/rickseven/logiq/internal/domain"
	"github.com/rickseven/logiq/internal/infrastructure/cache"
)

func checkCommand(cmdName string) string {
	_, err := exec.LookPath(cmdName)
	if err == nil {
		return "installed"
	}
	return "not found"
}

func fileExists(name string) bool {
	_, err := os.Stat(name)
	return err == nil
}

// GenerateDoctorResult captures the active environment configurations
func GenerateDoctorResult() domain.DoctorResult {
	if cached, found := cache.GetDoctor(); found {
		return *cached
	}

	res := domain.DoctorResult{
		Node:    checkCommand("node"),
		Npm:     checkCommand("npm"),
		Pnpm:    checkCommand("pnpm"),
		Flutter: checkCommand("flutter"),
		Dart:    checkCommand("dart"),
		Git:     checkCommand("git"),
		Vite:    "not found",
		Capabilities: []string{
			"**Shell commands** — install, build, test, lint, script apapun",
			"**Test runner output** — merangkum pass/fail count secara otomatis (Vitest/Jest/GoTest)",
			"**Build output** — deteksi sukses/gagal + jumlah module yang dikompilasi",
			"**Log analysis** — parse error traces, hitung errors & warnings dari raw text",
			"**Command explanation** — jelaskan command shell & chaining (&&, ;) dalam bahasa plain",
			"**Environment check** — verifikasi toolchain & kesehatan proyek",
		},
		Limitations: []string{
			"**Live Streaming** — Output hanya dirangkum SETELAH perintah selesai (MCP limitation)",
			"**Chaining Custom** — Masih optimal untuk tool standar, mungkin kurang akurat untuk piping yang sangat ekstrim",
		},
	}

	// Project Type Detection
	if fileExists("pubspec.yaml") {
		res.ProjectType = "Flutter"
	} else if fileExists("vite.config.js") || fileExists("vite.config.ts") {
		res.ProjectType = "Vue (Vite)"
		res.Vite = "detected"
	} else if fileExists("package.json") {
		res.ProjectType = "Node.js (Standard)"
	} else {
		res.ProjectType = "Unknown Environment"
	}

	if res.Vite == "not found" && checkCommand("vite") == "installed" {
		res.Vite = "installed"
	}

	cache.SetDoctor(res)
	return res
}
