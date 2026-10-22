// Package component is the parent of the UI's reusable building blocks.
//
// Every subdirectory is one independent component. A component:
//
//   - depends on nothing else in this project except ui/theme;
//   - never imports another component;
//   - contains no business logic, and knows nothing about skills, files or
//     the app's state;
//   - receives everything it needs as arguments. Sizes, titles and focus are
//     computed by the caller and passed in; a component never measures the
//     terminal or decides its own layout.
//
// Components are pure presentation. If one needs a piece of information, the
// caller supplies it.
package component
