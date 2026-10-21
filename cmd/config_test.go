package cmd

import (
	"reflect"
	"strings"
	"testing"

	"github.com/PiercingXX/devdash/internal"
	"github.com/spf13/viper"
)

// Each slice dimension represent row, col or widgets.
// For rows slice: [RowIndex][ColIndex][WidgetIndex]widget
// For size slice: [RowIndex][ColIndex]widget - no need the index of widgets
func Test_OrderWidgets(t *testing.T) {
	testCases := []struct {
		name          string
		project       Project
		expectedSizes [][]string
		expectedRows  [][][]internal.Widget
		wantErr       bool
	}{
		{
			name: "1 row 2 col 3 widgets",
			project: Project{
				Name: "test",
				Themes: map[string]map[string]string{
					"bar": {
						"border_color": "blue",
						"text_color":   "green",
					},
					"super_theme": {
						"border_color": "green",
						"text_color":   "blue",
					},
				},
				Widgets: []Row{
					{
						Row: []Column{
							{
								Col: []Widgets{
									{
										Size: "XL",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_1_widget1",
												Options: map[string]string{
													"start_date": "7_days_ago",
													"end_date":   "today",
												},
											},
										},
									},
									{
										Size: "L",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_1_widget2",
											},
										},
									},
								},
							},
							{
								Col: []Widgets{
									{
										Size: "M",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_2_widget1",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSizes: [][]string{
				{ // row 1
					"XL",
					"L",
					"M",
				},
			},
			expectedRows: [][][]internal.Widget{
				{ // row 1
					{ // column 1
						{
							Name: "row_1_col_1_widget1",
							Options: map[string]string{
								"start_date": "7_days_ago",
								"end_date":   "today",
							},
						},
						{
							Name: "row_1_col_1_widget2",
						},
					},
					{ //column 2
						{
							Name: "row_1_col_2_widget1",
						},
					},
				},
			},
		},
		{
			name: "3 row 3 col 5 widgets",
			project: Project{
				Name: "test",
				Widgets: []Row{
					{
						Row: []Column{
							{
								Col: []Widgets{
									{
										Size: "S",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_1_widget1",
											},
										},
									},
									{
										Size: "XS",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_1_widget2",
											},
										},
									},
								},
							},
							{
								Col: []Widgets{
									{
										Size: "M",
										Elements: []internal.Widget{
											{
												Name: "row_1_col_2_widget1",
											},
										},
									},
								},
							},
						},
					},
					{
						Row: []Column{
							{
								Col: []Widgets{
									{
										Size: "S",
										Elements: []internal.Widget{
											{
												Name: "row_2_col_1_widget1",
											},
										},
									},
								},
							},
							{
								Col: []Widgets{
									{
										Size: "L",
										Elements: []internal.Widget{
											{
												Name: "row_2_col_2_widget1",
											},
										},
									},
									{
										Size: "XL",
										Elements: []internal.Widget{
											{
												Name: "row_2_col_2_widget2",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expectedSizes: [][]string{
				{ // row 1
					"S",
					"XS",
					"M",
				},
				{
					"S",
					"L",
					"XL",
				},
			},
			expectedRows: [][][]internal.Widget{
				{ // row 1
					{ // column 1
						{
							Name: "row_1_col_1_widget1",
						},
						{
							Name: "row_1_col_1_widget2",
						},
					},
					{ //column 2
						{
							Name: "row_1_col_2_widget1",
						},
					},
				},
				{ // row 2
					{ // column 1
						{
							Name: "row_2_col_1_widget1",
						},
					},
					{ // column 1
						{
							Name: "row_2_col_2_widget1",
						},
						{
							Name: "row_2_col_2_widget2",
						},
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actualRows, actualSizes := tc.project.OrderWidgets()

			if tc.wantErr == false && !reflect.DeepEqual(actualRows, tc.expectedRows) {
				t.Errorf("Expected rows %v, actual %v", tc.expectedRows, actualRows)
			}

			if tc.wantErr == false && !reflect.DeepEqual(actualSizes, tc.expectedSizes) {
				t.Errorf("Expected sizes %v, actual %v", tc.expectedSizes, actualSizes)
			}

		})
	}
}

func Test_defaultConfig(t *testing.T) {
	v := viper.New()
	v.SetConfigType("yaml")
	if err := v.ReadConfig(strings.NewReader(defaultConfig("/home/me/.config/devdash"))); err != nil {
		t.Fatal(err)
	}

	var cfg config
	if err := v.Unmarshal(&cfg); err != nil {
		t.Fatal(err)
	}

	if len(cfg.Projects) != 1 || len(cfg.Projects[0].Widgets) == 0 {
		t.Fatalf("Expected 1 project with widgets, actual %+v", cfg.Projects)
	}
	if cfg.KEdit() != "C-e" || cfg.Projects[0].Services.Monitor.Address == "" {
		t.Errorf("Expected the edit key and the monitor service, actual %+v", cfg)
	}
}
