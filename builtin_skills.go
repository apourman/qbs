package qbs

import (
	"embed"
	"io/fs"
)

//go:embed all:.agents/skills
var builtinSkillFiles embed.FS

// BuiltinSkills returns the skill catalog shipped in this QBS release.
func BuiltinSkills() fs.FS {
	files, err := fs.Sub(builtinSkillFiles, ".agents/skills")
	if err != nil {
		panic(err)
	}
	return files
}
