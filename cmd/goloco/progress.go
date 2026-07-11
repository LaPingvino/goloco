package main

import (
	"encoding/json"
	"log"
	"os"
	"path/filepath"
	"time"
)

// Scenario completion tracking, persisted next to the binary as
// goloco_progress.json so the New Game menu can show completed scenarios.
type completedScenario struct {
	CompletedAt string `json:"completedAt"`
	Objective   string `json:"objective"`
	Score       uint32 `json:"score,omitempty"`
}

type progressFile struct {
	Completed map[string]completedScenario `json:"completed"` // key: scenario file base name
}

const progressPath = "goloco_progress.json"

func loadProgress() *progressFile {
	p := &progressFile{Completed: map[string]completedScenario{}}
	data, err := os.ReadFile(progressPath)
	if err != nil {
		return p
	}
	if err := json.Unmarshal(data, p); err != nil {
		log.Printf("[Progress] cannot parse %s: %v", progressPath, err)
	}
	if p.Completed == nil {
		p.Completed = map[string]completedScenario{}
	}
	return p
}

func (p *progressFile) save() {
	data, err := json.MarshalIndent(p, "", "  ")
	if err != nil {
		return
	}
	if err := os.WriteFile(progressPath, data, 0o644); err != nil {
		log.Printf("[Progress] cannot save: %v", err)
	}
}

// markCompleted records a win for the scenario at path and persists.
func (p *progressFile) markCompleted(path, objective string, score uint32) {
	name := filepath.Base(path)
	if _, done := p.Completed[name]; done {
		return
	}
	p.Completed[name] = completedScenario{
		CompletedAt: time.Now().Format("2006-01-02 15:04"),
		Objective:   objective,
		Score:       score,
	}
	p.save()
	log.Printf("[Progress] %s marked completed", name)
}

func (p *progressFile) isCompleted(path string) bool {
	_, done := p.Completed[filepath.Base(path)]
	return done
}
