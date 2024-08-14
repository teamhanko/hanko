package flowpilot

import (
	"testing"
)

// Test_defaultLink_Target tests the Target method of the defaultLink type.
func Test_defaultLink_Target(t *testing.T) {
	tests := []struct {
		name   string
		link   *defaultLink
		target LinkTarget
		want   LinkTarget
	}{
		{
			name:   "Set Target to _blank",
			link:   NewLink("Example", "Category1", "https://example.com").(*defaultLink),
			target: LinkTargetBlank,
			want:   LinkTargetBlank,
		},
		{
			name:   "Set Target to _self",
			link:   NewLink("Example", "Category1", "https://example.com").(*defaultLink),
			target: LinkTargetSelf,
			want:   LinkTargetSelf,
		},
		{
			name:   "Set Target to _top",
			link:   NewLink("Example", "Category1", "https://example.com").(*defaultLink),
			target: LinkTargetTop,
			want:   LinkTargetTop,
		},
		{
			name:   "Set Target to _parent",
			link:   NewLink("Example", "Category1", "https://example.com").(*defaultLink),
			target: LinkTargetParent,
			want:   LinkTargetParent,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.link.Target(tt.target).(*defaultLink).target
			if got != tt.want {
				t.Errorf("Target() = %v, want %v", got, tt.want)
			}
		})
	}
}

// Test_defaultLink_toResponseLink tests the toResponseLink method of the defaultLink type.
func Test_defaultLink_toResponseLink(t *testing.T) {
	tests := []struct {
		name string
		link *defaultLink
		want ResponseLink
	}{
		{
			name: "Convert link to ResponseLink",
			link: &defaultLink{
				name:     "Example",
				href:     "https://example.com",
				category: "Category1",
				target:   LinkTargetBlank,
			},
			want: ResponseLink{
				Name:     "Example",
				Href:     "https://example.com",
				Category: "Category1",
				Target:   LinkTargetBlank,
			},
		},
		{
			name: "Default link target",
			link: &defaultLink{
				name:     "Default Target",
				href:     "https://default.com",
				category: "Category2",
				target:   LinkTargetSelf, // Default target
			},
			want: ResponseLink{
				Name:     "Default Target",
				Href:     "https://default.com",
				Category: "Category2",
				Target:   LinkTargetSelf, // Default target
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.link.toResponseLink()
			if got != tt.want {
				t.Errorf("toResponseLink() = %v, want %v", got, tt.want)
			}
		})
	}
}
