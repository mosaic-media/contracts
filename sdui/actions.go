package sdui

// Action constructors. The Action uses pointer fields for optionals (so the JSON
// omits absent ones); these constructors hide that.

import (
	"encoding/json"

	"google.golang.org/protobuf/encoding/protojson"
)

func strp(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Navigate pushes another server-defined screen.
func Navigate(screen string, params map[string]any) Action {
	return Action{Kind: KindNavigate, Screen: strp(screen), Params: params}
}

// Back pops the client's navigation stack.
func Back() Action { return Action{Kind: KindBack} }

// OpenURL opens an external URL (the client validates the scheme).
func OpenURL(url string) Action { return Action{Kind: KindOpenURL, URL: strp(url)} }

// Invoke runs a Platform action by name — an intent the session transport
// dispatches (platform#37). The field is called Mutation for the same reason the
// contract's is: it names a Platform write, whatever the transport calls it.
func Invoke(mutation string, input map[string]any) Action {
	return Action{Kind: KindInvoke, Mutation: strp(mutation), Input: input}
}

// OpenOverlay presents a node as a modal/sheet/drawer. The node rides inside the
// action's open bag, so it is rendered to its JSON object form.
func OpenOverlay(surface Surface, node Node) Action {
	return Action{Kind: KindOpenOverlay, Surface: &surface, Node: nodeToMap(node)}
}

// CloseOverlay dismisses the topmost overlay.
func CloseOverlay() Action { return Action{Kind: KindCloseOverlay} }

// Play asks the client to resolve and play a content Part.
func Play(partID string) Action { return Action{Kind: KindPlayPart, PartID: strp(partID)} }

// Toast shows a transient message.
func Toast(message string, tone Tone) Action {
	a := Action{Kind: KindToast, Message: strp(message)}
	if tone != "" {
		a.Tone = &tone
	}
	return a
}

// Sequence runs several actions in order.
func Sequence(actions ...Action) Action {
	return Action{Kind: KindSequence, Actions: actions}
}

// Query re-reads a screen's data without changing the route — the refresh a
// search field wants when the term changes but the URL should not.
//
// A kind called `query` was removed from the contract by platform#37 along with
// GraphQL; that one carried a raw GraphQL string and a region name, and was
// unimplementable once there was no endpoint to send it to. This is a different
// action wearing the same name: it names a screen the server already knows how
// to build, exactly as Navigate does, and differs from Navigate only in not
// pushing history.
func Query(screen string, params map[string]any) Action {
	return Action{Kind: KindQuery, Screen: strp(screen), Params: params}
}

// SetValue writes a value into a named field of the enclosing state scope.
//
// Declared, and not yet interpreted by any client — the scopes it writes into
// arrive with state (S4) and forms (S5). It is in the vocabulary now because the
// vocabulary is versioned and negotiated as a whole, not because anything emits
// one; conformance/vocabulary.json is what makes that gap visible to a client
// rather than leaving it to be discovered by a control that does nothing.
func SetValue(field, value string) Action {
	return Action{Kind: KindSetValue, Field: strp(field), Value: strp(value)}
}

// Submit runs an action with the enclosing State scope's values merged into its
// input — what a form's button emits.
//
// It carries the action it runs, in the same field a Sequence carries its steps,
// rather than the enclosing scope carrying a "submitAction" the client has to go
// looking for. That is what makes a Form composed rather than special-cased: the
// scope stays a plain State, the button stays a plain Pressable, and the thing
// that knows about submission is the action, which is data.
//
// The merge is into the nested action's input, and the scope's values lose to
// anything the producer set explicitly there — a server that pinned a field is
// stating something the form must not overwrite.
//
// `into` names where in the action's input the values land — "settings" puts
// them at input.settings, and "" puts them at the top level. It exists because
// an input is rarely flat: a module's configureModule takes a whole settings
// document, and the field that was typed belongs inside it. Without a
// destination the collected values could only ever land at the top, which is
// why `$value` — a literal string substituted anywhere in the action — survived
// as long as it did.
//
// One path segment, not an expression. It is stated by the producer, resolved
// by no one, and the client does exactly one thing with it.
func Submit(action Action, into string) Action {
	a := Action{Kind: KindSubmit, Actions: []Action{action}}
	if into != "" {
		a.Field = strp(into)
	}
	return a
}

// nodeToMap renders a node to the JSON object form that rides inside an action's
// open bag (OpenOverlay). Nodes embedded in an action are rare; when absent this
// is never called.
func nodeToMap(node Node) map[string]any {
	if node == nil {
		return nil
	}
	b, err := protojson.Marshal(node)
	if err != nil {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		return nil
	}
	return m
}
