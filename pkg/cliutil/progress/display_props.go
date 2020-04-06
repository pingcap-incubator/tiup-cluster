package progress

// Mode determines how the progress bar is rendered
type Mode int

const (
	// ModeSpinner renders a Spinner
	ModeSpinner Mode = iota
	// ModeProgress renders a ProgressBar. Not supported yet.
	ModeProgress
	// ModeDone renders as "Done" message.
	ModeDone
	// ModeError renders as "Error" message.
	ModeError
)

// DisplayProps controls the display of the progress bar.
type DisplayProps struct {
	Prefix string
	Suffix string // If `Mode == Done / Error`, Suffix is not printed
	Mode   Mode
}
