package terma

import "testing"

func TestSnapshot_Breadcrumbs_Basic(t *testing.T) {
	widget := Breadcrumbs{
		ID:   "breadcrumbs-basic",
		Path: []string{"Commands", "File", "Recent"},
	}
	AssertSnapshot(t, widget, 40, 3, "Breadcrumbs with three segments separated by >")
}
