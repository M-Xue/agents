// Package skills describes the layout of the agents workspace.
//
// It is the one place that knows where things live: the directory that
// exposes selected skills to the agent, the directory of skill
// libraries, where a library keeps its skills and what a skill's main file
// is called. It holds no behaviour beyond building those paths and locating
// the workspace root; discovering, reading and linking skills is done by app
// using files.
package skills
