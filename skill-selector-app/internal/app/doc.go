// Package app holds the general business logic of skill-selector.
//
// It is the only package that joins skills (where things live) with files
// (how to touch them). It loads the libraries and their skills from disk,
// answers which skills are exposed, and links or unlinks a skill when it is
// toggled. The ui package drives it and displays the results; it knows
// nothing about the terminal.
package app
