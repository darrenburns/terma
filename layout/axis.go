package layout

// Axis represents the primary direction of a linear layout.
type Axis int

const (
	// Horizontal axis: main-axis runs left-to-right (X), cross-axis runs top-to-bottom (Y).
	// Used by RowNode.
	Horizontal Axis = iota

	// Vertical axis: main-axis runs top-to-bottom (Y), cross-axis runs left-to-right (X).
	// Used by ColumnNode.
	Vertical
)

// MainAxisAlignment controls how children are distributed along the main axis
// when there is extra space available.
type MainAxisAlignment int

const (
	// MainAxisStart packs children at the start of the main axis.
	MainAxisStart MainAxisAlignment = iota

	// MainAxisCenter centers children along the main axis.
	MainAxisCenter

	// MainAxisEnd packs children at the end of the main axis.
	MainAxisEnd

	// MainAxisSpaceBetween distributes extra space evenly between children.
	// First child at start, last child at end, equal gaps between.
	MainAxisSpaceBetween

	// MainAxisSpaceAround distributes extra space evenly around children.
	// Each child gets equal space on both sides (gaps between children are 2x edge gaps).
	MainAxisSpaceAround

	// MainAxisSpaceEvenly distributes extra space so all gaps (including edges) are equal.
	MainAxisSpaceEvenly
)

// CrossAxisAlignment controls how children are positioned along the cross axis.
type CrossAxisAlignment int

const (
	// CrossAxisStart aligns children at the start of the cross axis.
	CrossAxisStart CrossAxisAlignment = iota

	// CrossAxisCenter centers children along the cross axis.
	CrossAxisCenter

	// CrossAxisEnd aligns children at the end of the cross axis.
	CrossAxisEnd

	// CrossAxisStretch stretches children to fill the cross axis.
	// Children are re-laid out with tight cross-axis constraints.
	CrossAxisStretch
)
