// Package files wraps filesystem access.
//
// It exposes exactly the operations the rest of the program needs and
// nothing else: reading a file as text, listing the directories under a
// path, and adding, removing and checking symlinks that expose one folder
// inside another. It contains no business logic and knows nothing about
// skills.
//
// Every symlink operation takes the same two arguments: dir, the folder the
// link lives in, and target, the folder being linked. The link is always
// named after the last path element of target, so linking
// /lib/skills/code-review into /exposed creates /exposed/code-review.
package files
