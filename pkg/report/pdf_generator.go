/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package report

import (
	"bytes"
	"fmt"
	"html"
	"sort"
	"strings"
	"time"

	"github.com/jung-kurt/gofpdf"

	assessmentv1alpha1 "github.com/openshift-assessment/cluster-assessment-operator/api/v1alpha1"
)

// Colors for status badges
var (
	colorPass = []int{34, 139, 34}  // Forest Green
	colorWarn = []int{255, 165, 0}  // Orange
	colorFail = []int{220, 20, 60}  // Crimson
	colorInfo = []int{70, 130, 180} // Steel Blue
)

// colorForStatus returns the color palette for a given FindingStatus.
func colorForStatus(status assessmentv1alpha1.FindingStatus) []int {
	switch status {
	case assessmentv1alpha1.FindingStatusPass:
		return colorPass
	case assessmentv1alpha1.FindingStatusWarn:
		return colorWarn
	case assessmentv1alpha1.FindingStatusFail:
		return colorFail
	case assessmentv1alpha1.FindingStatusInfo:
		return colorInfo
	default:
		return colorInfo
	}
}

// labelForStatus returns the display label for a given FindingStatus.
func labelForStatus(status assessmentv1alpha1.FindingStatus) string {
	switch status {
	case assessmentv1alpha1.FindingStatusPass:
		return "PASS"
	case assessmentv1alpha1.FindingStatusWarn:
		return "WARNING"
	case assessmentv1alpha1.FindingStatusFail:
		return "FAILED"
	case assessmentv1alpha1.FindingStatusInfo:
		return "INFO"
	default:
		return string(status)
	}
}

// pageWidth returns the usable content width (A4 minus margins).
const (
	pageContentWidth = 180.0 // A4 width (210mm) - 15mm margins on each side
	leftMargin       = 15.0
)

// GeneratePDF creates a professional PDF report from the assessment.
func GeneratePDF(assessment *assessmentv1alpha1.ClusterAssessment) ([]byte, error) {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(leftMargin, 15, 15)

	// Register footer with page numbers
	pdf.SetFooterFunc(func() {
		pdf.SetY(-15)
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(150, 150, 150)
		pdf.CellFormat(0, 10,
			fmt.Sprintf("OpenShift Cluster Assessment Report  |  %s  |  Page %d/{nb}",
				assessment.Status.ClusterInfo.ClusterID, pdf.PageNo()),
			"", 0, "C", false, 0, "")
	})
	pdf.AliasNbPages("")

	// --- Cover Page ---
	addCoverPage(pdf, assessment)

	// --- Content Pages ---
	pdf.AddPage()

	// Cluster Info Box
	addSectionTitle(pdf, "Cluster Information")
	addClusterInfoTable(pdf, assessment)
	pdf.Ln(10)

	// Summary Section
	addSectionTitle(pdf, "Assessment Summary")
	addSummarySection(pdf, assessment)
	pdf.Ln(10)

	// Score visualization
	if assessment.Status.Summary.Score != nil {
		addScoreVisualization(pdf, *assessment.Status.Summary.Score)
		pdf.Ln(10)
	}

	// Delta Section (changes since last run)
	if assessment.Status.Delta != nil {
		addDeltaSection(pdf, assessment)
		pdf.Ln(10)
	}

	// Findings by Category (horizontal bar chart)
	addSectionTitle(pdf, "Findings by Category")
	addCategoryBarChart(pdf, assessment)
	pdf.Ln(5)

	// Detailed Findings
	pdf.AddPage()
	addSectionTitle(pdf, "Detailed Findings")
	addDetailedFindings(pdf, assessment)

	if err := pdf.Error(); err != nil {
		return nil, fmt.Errorf("PDF generation error: %w", err)
	}

	// Output to bytes
	var buf bytes.Buffer
	if err := pdf.Output(&buf); err != nil {
		return nil, fmt.Errorf("failed to generate PDF: %w", err)
	}

	return buf.Bytes(), nil
}

// addCoverPage renders a professional cover page.
func addCoverPage(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	pdf.AddPage()

	// Top accent bar
	pdf.SetFillColor(0, 51, 102)
	pdf.Rect(0, 0, 210, 8, "F")

	// Main title area
	pdf.SetY(60)
	pdf.SetFont("Helvetica", "B", 32)
	pdf.SetTextColor(0, 51, 102)
	pdf.CellFormat(0, 15, "OpenShift Cluster", "", 1, "C", false, 0, "")
	pdf.CellFormat(0, 15, "Assessment Report", "", 1, "C", false, 0, "")
	pdf.Ln(10)

	// Horizontal rule
	pdf.SetDrawColor(0, 51, 102)
	pdf.SetLineWidth(0.8)
	pdf.Line(50, pdf.GetY(), 160, pdf.GetY())
	pdf.Ln(12)

	// Cluster info on cover
	pdf.SetFont("Helvetica", "", 14)
	pdf.SetTextColor(80, 80, 80)
	info := assessment.Status.ClusterInfo
	if info.ClusterID != "" {
		pdf.CellFormat(0, 8, fmt.Sprintf("Cluster: %s", info.ClusterID), "", 1, "C", false, 0, "")
	}
	if info.ClusterVersion != "" {
		pdf.CellFormat(0, 8, fmt.Sprintf("OpenShift %s  |  %s", info.ClusterVersion, info.Platform), "", 1, "C", false, 0, "")
	}
	pdf.Ln(5)

	// Date
	pdf.SetFont("Helvetica", "", 12)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 8, fmt.Sprintf("Generated: %s", time.Now().Format("January 2, 2006 at 15:04 MST")), "", 1, "C", false, 0, "")
	pdf.Ln(15)

	// Score circle (large, centered)
	if assessment.Status.Summary.Score != nil {
		score := *assessment.Status.Summary.Score
		centerX := 105.0
		centerY := pdf.GetY() + 25.0
		radius := 22.0

		// Circle background
		color := colorForStatus(assessmentv1alpha1.FindingStatusPass)
		if score < 60 {
			color = colorForStatus(assessmentv1alpha1.FindingStatusFail)
		} else if score < 80 {
			color = colorForStatus(assessmentv1alpha1.FindingStatusWarn)
		}
		pdf.SetFillColor(color[0], color[1], color[2])
		pdf.Circle(centerX, centerY, radius, "F")

		// Score text
		pdf.SetFont("Helvetica", "B", 28)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetXY(centerX-radius, centerY-8)
		pdf.CellFormat(radius*2, 16, fmt.Sprintf("%d%%", score), "", 1, "C", false, 0, "")

		// Label
		pdf.SetFont("Helvetica", "", 10)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetXY(centerX-radius, centerY+5)
		pdf.CellFormat(radius*2, 6, "Overall Score", "", 1, "C", false, 0, "")

		pdf.SetY(centerY + radius + 10)
	}

	// Summary counts on cover
	summary := assessment.Status.Summary
	pdf.SetY(pdf.GetY() + 5)
	boxWidth := 35.0
	totalWidth := boxWidth*4 + 5*3
	startX := (210 - totalWidth) / 2
	y := pdf.GetY()

	summaryItems := []struct {
		label string
		count int
		color []int
	}{
		{"PASS", summary.PassCount, colorPass},
		{"WARN", summary.WarnCount, colorWarn},
		{"FAIL", summary.FailCount, colorFail},
		{"INFO", summary.InfoCount, colorInfo},
	}

	for i, item := range summaryItems {
		x := startX + float64(i)*(boxWidth+5)
		pdf.SetFillColor(item.color[0], item.color[1], item.color[2])
		pdf.RoundedRect(x, y, boxWidth, 18, 3, "1234", "F")

		pdf.SetFont("Helvetica", "B", 14)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetXY(x, y+1)
		pdf.CellFormat(boxWidth, 10, fmt.Sprintf("%d", item.count), "", 0, "C", false, 0, "")

		pdf.SetFont("Helvetica", "", 8)
		pdf.SetXY(x, y+11)
		pdf.CellFormat(boxWidth, 6, item.label, "", 0, "C", false, 0, "")
	}

	// Profile used
	pdf.SetY(y + 30)
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(120, 120, 120)
	profileUsed := assessment.Status.Summary.ProfileUsed
	if profileUsed == "" {
		profileUsed = assessment.Spec.Profile
	}
	pdf.CellFormat(0, 6, fmt.Sprintf("Profile: %s  |  Total Checks: %d", profileUsed, summary.TotalChecks), "", 1, "C", false, 0, "")

	// Bottom accent bar
	pdf.SetFillColor(0, 51, 102)
	pdf.Rect(0, 289, 210, 8, "F")
}

func addSectionTitle(pdf *gofpdf.Fpdf, title string) {
	pdf.SetFont("Helvetica", "B", 14)
	pdf.SetTextColor(0, 51, 102)
	pdf.SetFillColor(240, 240, 245)
	pdf.CellFormat(0, 10, title, "", 1, "L", true, 0, "")
	pdf.Ln(3)
}

func addClusterInfoTable(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	pdf.SetFont("Helvetica", "", 10)
	pdf.SetTextColor(0, 0, 0)

	info := assessment.Status.ClusterInfo

	// Two column layout
	colWidth := 85.0
	rowHeight := 7.0

	profileUsed := assessment.Status.Summary.ProfileUsed
	if profileUsed == "" {
		profileUsed = assessment.Spec.Profile
	}

	rows := [][]string{
		{"Cluster ID:", info.ClusterID},
		{"OpenShift Version:", info.ClusterVersion},
		{"Platform:", info.Platform},
		{"Update Channel:", info.Channel},
		{"Total Nodes:", fmt.Sprintf("%d", info.NodeCount)},
		{"Control Plane Nodes:", fmt.Sprintf("%d", info.ControlPlaneNodes)},
		{"Worker Nodes:", fmt.Sprintf("%d", info.WorkerNodes)},
		{"Assessment Profile:", profileUsed},
	}

	for _, row := range rows {
		pdf.SetFont("Helvetica", "B", 10)
		pdf.CellFormat(colWidth, rowHeight, row[0], "", 0, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 10)
		pdf.CellFormat(colWidth, rowHeight, row[1], "", 1, "L", false, 0, "")
	}
}

func addSummarySection(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	summary := assessment.Status.Summary

	// Summary boxes
	boxWidth := 40.0
	boxHeight := 20.0
	startX := leftMargin
	y := pdf.GetY()

	summaryItems := []struct {
		label string
		count int
		color []int
	}{
		{"PASS", summary.PassCount, colorPass},
		{"WARN", summary.WarnCount, colorWarn},
		{"FAIL", summary.FailCount, colorFail},
		{"INFO", summary.InfoCount, colorInfo},
	}

	for i, item := range summaryItems {
		x := startX + float64(i)*(boxWidth+5)

		// Box background
		pdf.SetFillColor(item.color[0], item.color[1], item.color[2])
		pdf.RoundedRect(x, y, boxWidth, boxHeight, 3, "1234", "F")

		// Count
		pdf.SetFont("Helvetica", "B", 16)
		pdf.SetTextColor(255, 255, 255)
		pdf.SetXY(x, y+2)
		pdf.CellFormat(boxWidth, 10, fmt.Sprintf("%d", item.count), "", 0, "C", false, 0, "")

		// Label
		pdf.SetFont("Helvetica", "", 9)
		pdf.SetXY(x, y+12)
		pdf.CellFormat(boxWidth, 6, item.label, "", 0, "C", false, 0, "")
	}

	pdf.SetY(y + boxHeight + 5)
	pdf.SetTextColor(0, 0, 0)

	// Total checks
	pdf.SetFont("Helvetica", "", 10)
	pdf.CellFormat(0, 6, fmt.Sprintf("Total Checks: %d", summary.TotalChecks), "", 1, "L", false, 0, "")
}

func addScoreVisualization(pdf *gofpdf.Fpdf, score int) {
	y := pdf.GetY()

	// Score label
	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(30, 10, "Score:", "", 0, "L", false, 0, "")

	// Progress bar background
	barWidth := 120.0
	barHeight := 10.0
	barX := 45.0

	pdf.SetFillColor(220, 220, 220)
	pdf.RoundedRect(barX, y, barWidth, barHeight, 2, "1234", "F")

	// Progress bar fill
	fillWidth := barWidth * float64(score) / 100.0
	color := colorForStatus(assessmentv1alpha1.FindingStatusPass)
	if score < 60 {
		color = colorForStatus(assessmentv1alpha1.FindingStatusFail)
	} else if score < 80 {
		color = colorForStatus(assessmentv1alpha1.FindingStatusWarn)
	}
	pdf.SetFillColor(color[0], color[1], color[2])
	if fillWidth > 0 {
		pdf.RoundedRect(barX, y, fillWidth, barHeight, 2, "1234", "F")
	}

	// Score text
	pdf.SetFont("Helvetica", "B", 11)
	pdf.SetTextColor(255, 255, 255)
	pdf.SetXY(barX, y)
	pdf.CellFormat(barWidth, barHeight, fmt.Sprintf("%d%%", score), "", 0, "C", false, 0, "")

	pdf.SetY(y + barHeight + 2)
}

// addDeltaSection renders a section showing changes since the last assessment run.
func addDeltaSection(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	delta := assessment.Status.Delta
	if delta == nil {
		return
	}

	addSectionTitle(pdf, "Changes Since Last Run")

	y := pdf.GetY()

	// Score delta
	if delta.ScoreDelta != nil && *delta.ScoreDelta != 0 {
		pdf.SetFont("Helvetica", "B", 12)
		scoreDelta := *delta.ScoreDelta
		if scoreDelta > 0 {
			pdf.SetTextColor(colorPass[0], colorPass[1], colorPass[2])
			pdf.CellFormat(0, 8, fmt.Sprintf("Score: +%d points (improved)", scoreDelta), "", 1, "L", false, 0, "")
		} else {
			pdf.SetTextColor(colorFail[0], colorFail[1], colorFail[2])
			pdf.CellFormat(0, 8, fmt.Sprintf("Score: %d points (regressed)", scoreDelta), "", 1, "L", false, 0, "")
		}
		pdf.Ln(3)
	}

	// Delta summary boxes
	type deltaItem struct {
		label string
		items []string
		color []int
		icon  string
	}

	deltaItems := []deltaItem{
		{"New Issues", delta.NewFindings, colorFail, "+"},
		{"Resolved", delta.ResolvedFindings, colorPass, "-"},
		{"Regressions", delta.RegressionFindings, colorWarn, "!"},
		{"Improved", delta.ImprovedFindings, colorInfo, "*"},
	}

	// Summary row
	boxWidth := 42.0
	boxHeight := 14.0
	y = pdf.GetY()
	for i, item := range deltaItems {
		x := leftMargin + float64(i)*(boxWidth+3)

		// Light background with colored left border
		pdf.SetFillColor(248, 248, 250)
		pdf.RoundedRect(x, y, boxWidth, boxHeight, 2, "1234", "F")
		pdf.SetFillColor(item.color[0], item.color[1], item.color[2])
		pdf.Rect(x, y, 3, boxHeight, "F")

		// Count
		pdf.SetFont("Helvetica", "B", 12)
		pdf.SetTextColor(item.color[0], item.color[1], item.color[2])
		pdf.SetXY(x+5, y+1)
		pdf.CellFormat(15, 6, fmt.Sprintf("%s%d", item.icon, len(item.items)), "", 0, "L", false, 0, "")

		// Label
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(80, 80, 80)
		pdf.SetXY(x+5, y+7)
		pdf.CellFormat(boxWidth-5, 5, item.label, "", 0, "L", false, 0, "")
	}

	pdf.SetY(y + boxHeight + 4)

	// List finding IDs if any
	pdf.SetTextColor(0, 0, 0)
	for _, item := range deltaItems {
		if len(item.items) == 0 {
			continue
		}
		pdf.SetFont("Helvetica", "B", 8)
		pdf.SetTextColor(item.color[0], item.color[1], item.color[2])
		pdf.CellFormat(0, 5, fmt.Sprintf("%s:", item.label), "", 1, "L", false, 0, "")

		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(80, 80, 80)
		// Show up to 10 finding IDs per line
		for i := 0; i < len(item.items); i += 10 {
			end := i + 10
			if end > len(item.items) {
				end = len(item.items)
			}
			line := strings.Join(item.items[i:end], ", ")
			pdf.CellFormat(0, 4, "  "+line, "", 1, "L", false, 0, "")
		}
		pdf.Ln(1)
	}
}

// addCategoryBarChart renders a horizontal stacked bar chart for each category.
func addCategoryBarChart(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	// Group findings by category
	type categoryCounts struct {
		pass, warn, fail, info int
		total                  int
	}
	categories := make(map[string]*categoryCounts)
	for _, f := range assessment.Status.Findings {
		c, ok := categories[f.Category]
		if !ok {
			c = &categoryCounts{}
			categories[f.Category] = c
		}
		c.total++
		switch f.Status {
		case assessmentv1alpha1.FindingStatusPass:
			c.pass++
		case assessmentv1alpha1.FindingStatusWarn:
			c.warn++
		case assessmentv1alpha1.FindingStatusFail:
			c.fail++
		case assessmentv1alpha1.FindingStatusInfo:
			c.info++
		}
	}

	// Sort category names for deterministic output
	sortedNames := make([]string, 0, len(categories))
	for name := range categories {
		sortedNames = append(sortedNames, name)
	}
	sort.Strings(sortedNames)

	// Find max total for scaling
	maxTotal := 0
	for _, c := range categories {
		if c.total > maxTotal {
			maxTotal = c.total
		}
	}
	if maxTotal == 0 {
		return
	}

	labelWidth := 55.0
	barMaxWidth := pageContentWidth - labelWidth - 30 // leave room for count label
	rowHeight := 10.0

	pdf.SetFont("Helvetica", "", 9)
	pdf.SetTextColor(0, 0, 0)

	for _, name := range sortedNames {
		c := categories[name]

		if pdf.GetY() > 260 {
			pdf.AddPage()
		}

		y := pdf.GetY()

		// Category label
		pdf.SetFont("Helvetica", "B", 9)
		pdf.SetTextColor(50, 50, 50)
		pdf.SetXY(leftMargin, y)
		pdf.CellFormat(labelWidth, rowHeight, name, "", 0, "R", false, 0, "")

		// Stacked bar
		barX := leftMargin + labelWidth + 3
		scale := barMaxWidth / float64(maxTotal)

		segments := []struct {
			count int
			color []int
		}{
			{c.fail, colorFail},
			{c.warn, colorWarn},
			{c.info, colorInfo},
			{c.pass, colorPass},
		}

		currentX := barX
		for _, seg := range segments {
			if seg.count == 0 {
				continue
			}
			segWidth := float64(seg.count) * scale
			if segWidth < 1 {
				segWidth = 1
			}
			pdf.SetFillColor(seg.color[0], seg.color[1], seg.color[2])
			pdf.Rect(currentX, y+1, segWidth, rowHeight-2, "F")
			currentX += segWidth
		}

		// Total count label
		pdf.SetFont("Helvetica", "", 8)
		pdf.SetTextColor(100, 100, 100)
		pdf.SetXY(currentX+2, y)
		pdf.CellFormat(25, rowHeight, fmt.Sprintf("%d checks", c.total), "", 0, "L", false, 0, "")

		pdf.SetY(y + rowHeight + 1)
	}

	// Legend
	pdf.Ln(3)
	legendY := pdf.GetY()
	legendItems := []struct {
		label string
		color []int
	}{
		{"Fail", colorFail},
		{"Warn", colorWarn},
		{"Info", colorInfo},
		{"Pass", colorPass},
	}
	legendX := leftMargin + labelWidth + 3
	for _, item := range legendItems {
		pdf.SetFillColor(item.color[0], item.color[1], item.color[2])
		pdf.Rect(legendX, legendY+1, 6, 4, "F")
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(80, 80, 80)
		pdf.SetXY(legendX+7, legendY)
		pdf.CellFormat(20, 6, item.label, "", 0, "L", false, 0, "")
		legendX += 28
	}
	pdf.SetY(legendY + 8)
}

func addDetailedFindings(pdf *gofpdf.Fpdf, assessment *assessmentv1alpha1.ClusterAssessment) {
	// Group findings by status for better organization
	statusOrder := []assessmentv1alpha1.FindingStatus{
		assessmentv1alpha1.FindingStatusFail,
		assessmentv1alpha1.FindingStatusWarn,
		assessmentv1alpha1.FindingStatusInfo,
		assessmentv1alpha1.FindingStatusPass,
	}

	// Group findings by status in a single pass
	findingsByStatus := make(map[assessmentv1alpha1.FindingStatus][]assessmentv1alpha1.Finding)
	for _, f := range assessment.Status.Findings {
		findingsByStatus[f.Status] = append(findingsByStatus[f.Status], f)
	}

	for _, status := range statusOrder {
		findings := findingsByStatus[status]
		if len(findings) == 0 {
			continue
		}

		// Status header
		addStatusHeader(pdf, status, len(findings))

		for _, f := range findings {
			addFindingCard(pdf, f)
		}
		pdf.Ln(5)
	}
}

func addStatusHeader(pdf *gofpdf.Fpdf, status assessmentv1alpha1.FindingStatus, count int) {
	color := colorForStatus(status)
	label := labelForStatus(status)

	pdf.SetFont("Helvetica", "B", 12)
	pdf.SetTextColor(color[0], color[1], color[2])
	pdf.CellFormat(0, 8, fmt.Sprintf("%s (%d)", label, count), "", 1, "L", false, 0, "")
	pdf.SetTextColor(0, 0, 0)
}

// addFindingCard renders a single finding card with dynamically calculated height.
func addFindingCard(pdf *gofpdf.Fpdf, f assessmentv1alpha1.Finding) {
	// Calculate all content lines first to determine card height
	title := f.Title
	description := f.Description
	hasRecommendation := (f.Status == assessmentv1alpha1.FindingStatusFail || f.Status == assessmentv1alpha1.FindingStatusWarn) && f.Recommendation != ""
	hasRemediation := f.Remediation != nil && len(f.Remediation.Commands) > 0
	hasReferences := len(f.References) > 0
	hasImpact := f.Impact != ""
	hasResource := f.Resource != ""

	// Estimate card height dynamically
	cardHeight := 8.0 // title line height

	// Description height: estimate lines wrapped at ~165mm width with 8pt font
	descLines := estimateWrappedLines(description, 165, 8)
	cardHeight += float64(descLines) * 4.0

	// Metadata line (category + validator + resource)
	cardHeight += 5.0

	// Impact
	if hasImpact {
		impactLines := estimateWrappedLines(f.Impact, 165, 8)
		cardHeight += float64(impactLines)*4.0 + 3.0
	}

	// Recommendation
	if hasRecommendation {
		recLines := estimateWrappedLines("Recommendation: "+f.Recommendation, 176, 8)
		cardHeight += float64(recLines)*4.0 + 6.0
	}

	// References
	if hasReferences {
		cardHeight += 5.0
	}

	// Remediation
	remediationHeight := 0.0
	if hasRemediation {
		remediationHeight += 6.0 // safety label
		if f.Remediation.EstimatedImpact != "" {
			remediationHeight += 4.0
		}
		if len(f.Remediation.Prerequisites) > 0 {
			remediationHeight += 5.0 + float64(len(f.Remediation.Prerequisites))*4.0
		}
		for _, cmd := range f.Remediation.Commands {
			if cmd.Description != "" {
				remediationHeight += 4.0
			}
			remediationHeight += 5.0 // command line
		}
		if f.Remediation.DocumentationURL != "" {
			remediationHeight += 5.0
		}
		remediationHeight += 4.0 // padding
	}

	totalHeight := cardHeight + 6 // 6 for padding around card
	if totalHeight < 28 {
		totalHeight = 28
	}

	// Check if we need a new page (account for card + remediation)
	if pdf.GetY()+totalHeight+remediationHeight > 270 {
		pdf.AddPage()
	}

	startY := pdf.GetY()

	// Card background
	pdf.SetFillColor(248, 248, 250)
	pdf.RoundedRect(leftMargin, startY, pageContentWidth, totalHeight, 2, "1234", "F")

	// Status badge (colored indicator)
	color := colorForStatus(f.Status)
	pdf.SetFillColor(color[0], color[1], color[2])
	pdf.RoundedRect(leftMargin+2, startY+2, 8, 8, 1, "1234", "F")

	// Title
	currentY := startY + 2
	pdf.SetXY(leftMargin+13, currentY)
	pdf.SetFont("Helvetica", "B", 10)
	pdf.SetTextColor(0, 0, 0)
	pdf.CellFormat(pageContentWidth-15, 5, title, "", 1, "L", false, 0, "")
	currentY += 7

	// Description (word-wrapped)
	pdf.SetXY(leftMargin+13, currentY)
	pdf.SetFont("Helvetica", "", 8)
	pdf.SetTextColor(80, 80, 80)
	pdf.MultiCell(pageContentWidth-15, 4, description, "", "L", false)
	currentY = pdf.GetY() + 1

	// Resource/Namespace (if present)
	if hasResource {
		pdf.SetXY(leftMargin+13, currentY)
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(100, 100, 100)
		resourceStr := "Resource: " + f.Resource
		if f.Namespace != "" {
			resourceStr += " (ns: " + f.Namespace + ")"
		}
		pdf.CellFormat(0, 4, resourceStr, "", 1, "L", false, 0, "")
		currentY += 5
	}

	// Category and Validator
	pdf.SetXY(leftMargin+13, currentY)
	pdf.SetFont("Helvetica", "", 7)
	pdf.SetTextColor(120, 120, 120)
	pdf.CellFormat(0, 4, fmt.Sprintf("Category: %s  |  Validator: %s", f.Category, f.Validator), "", 1, "L", false, 0, "")
	currentY += 5

	// Impact (if present)
	if hasImpact {
		pdf.SetXY(leftMargin+13, currentY)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(90, 70, 50)
		pdf.MultiCell(pageContentWidth-15, 4, "Impact: "+f.Impact, "", "L", false)
		currentY = pdf.GetY() + 1
	}

	// Set Y past the card
	if currentY > startY+totalHeight {
		pdf.SetY(currentY + 2)
	} else {
		pdf.SetY(startY + totalHeight + 2)
	}

	// Recommendation box (outside main card)
	if hasRecommendation {
		recY := pdf.GetY()
		pdf.SetFillColor(255, 250, 240)
		pdf.SetXY(leftMargin+5, recY)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(100, 80, 60)
		pdf.MultiCell(pageContentWidth-10, 4, "Recommendation: "+f.Recommendation, "", "L", false)
		recEndY := pdf.GetY()
		// Draw background behind the recommendation (go back and fill)
		pdf.SetFillColor(255, 250, 240)
		pdf.RoundedRect(leftMargin, recY-1, pageContentWidth, recEndY-recY+2, 2, "1234", "F")
		// Redraw text on top of background
		pdf.SetXY(leftMargin+5, recY)
		pdf.SetFont("Helvetica", "I", 8)
		pdf.SetTextColor(100, 80, 60)
		pdf.MultiCell(pageContentWidth-10, 4, "Recommendation: "+f.Recommendation, "", "L", false)
		pdf.Ln(1)
	}

	// References
	if hasReferences {
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(70, 130, 180)
		refs := make([]string, 0, len(f.References))
		for _, ref := range f.References {
			if len(ref) > 80 {
				refs = append(refs, ref[:77]+"...")
			} else {
				refs = append(refs, ref)
			}
		}
		pdf.CellFormat(0, 4, "Refs: "+strings.Join(refs, " | "), "", 1, "L", false, 0, "")
		pdf.Ln(1)
	}

	// Remediation section
	if hasRemediation {
		addRemediationBlock(pdf, f.Remediation)
	}

	pdf.Ln(3)
}

// addRemediationBlock renders the structured remediation guidance for a finding.
func addRemediationBlock(pdf *gofpdf.Fpdf, rem *assessmentv1alpha1.RemediationGuidance) {
	if pdf.GetY() > 255 {
		pdf.AddPage()
	}

	// Safety label
	pdf.SetFont("Helvetica", "B", 8)
	safetyColor := colorForStatus(assessmentv1alpha1.FindingStatusInfo)
	switch rem.Safety {
	case assessmentv1alpha1.RemediationSafeApply:
		safetyColor = colorPass
	case assessmentv1alpha1.RemediationRequiresReview:
		safetyColor = colorWarn
	case assessmentv1alpha1.RemediationDestructive:
		safetyColor = colorFail
	}
	pdf.SetTextColor(safetyColor[0], safetyColor[1], safetyColor[2])
	pdf.CellFormat(0, 4, fmt.Sprintf("Remediation [%s]:", rem.Safety), "", 1, "L", false, 0, "")

	// Estimated impact
	if rem.EstimatedImpact != "" {
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(0, 4, "  Impact: "+rem.EstimatedImpact, "", 1, "L", false, 0, "")
	}

	// Prerequisites
	if len(rem.Prerequisites) > 0 {
		pdf.SetFont("Helvetica", "B", 7)
		pdf.SetTextColor(80, 80, 80)
		pdf.CellFormat(0, 4, "  Prerequisites:", "", 1, "L", false, 0, "")
		pdf.SetFont("Helvetica", "", 7)
		for _, prereq := range rem.Prerequisites {
			if pdf.GetY() > 270 {
				pdf.AddPage()
			}
			pdf.CellFormat(0, 4, "    - "+prereq, "", 1, "L", false, 0, "")
		}
	}

	// Commands
	for _, cmd := range rem.Commands {
		if pdf.GetY() > 270 {
			pdf.AddPage()
		}
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(80, 80, 80)
		if cmd.Description != "" {
			prefix := ""
			if cmd.RequiresConfirmation {
				prefix = "[!] "
			}
			pdf.CellFormat(0, 3, "  "+prefix+cmd.Description, "", 1, "L", false, 0, "")
		}

		// Command in monospace with background
		cmdY := pdf.GetY()
		pdf.SetFillColor(40, 40, 50)
		pdf.Rect(leftMargin+5, cmdY, pageContentWidth-10, 5, "F")
		pdf.SetFont("Courier", "", 7)
		pdf.SetTextColor(200, 210, 220)
		pdf.SetXY(leftMargin+7, cmdY+1)
		pdf.CellFormat(pageContentWidth-14, 3, "$ "+cmd.Command, "", 1, "L", false, 0, "")
		pdf.SetY(cmdY + 6)
	}

	// Documentation URL
	if rem.DocumentationURL != "" {
		pdf.SetFont("Helvetica", "", 7)
		pdf.SetTextColor(70, 130, 180)
		docURL := rem.DocumentationURL
		if len(docURL) > 90 {
			docURL = docURL[:87] + "..."
		}
		pdf.CellFormat(0, 4, "  Docs: "+docURL, "", 1, "L", false, 0, "")
	}

	pdf.Ln(1)
}

// estimateWrappedLines estimates how many lines text will take when wrapped at a given width.
func estimateWrappedLines(text string, widthMM float64, fontSizePt float64) int {
	if text == "" {
		return 0
	}
	// Approximate: Helvetica 8pt ≈ 2.1mm per char, 10pt ≈ 2.6mm per char
	charWidth := fontSizePt * 0.26 // rough mm per char
	charsPerLine := int(widthMM / charWidth)
	if charsPerLine < 1 {
		charsPerLine = 1
	}
	lines := (len(text) + charsPerLine - 1) / charsPerLine
	if lines < 1 {
		lines = 1
	}
	return lines
}

// GenerateHTML creates an HTML report that can be easily converted to PDF.
func GenerateHTML(assessment *assessmentv1alpha1.ClusterAssessment) ([]byte, error) {
	var buf bytes.Buffer

	buf.WriteString(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8">
    <title>OpenShift Cluster Assessment Report</title>
    <style>
        body { font-family: 'Segoe UI', Arial, sans-serif; margin: 40px; background: #f5f5f5; }
        .container { max-width: 900px; margin: 0 auto; background: white; padding: 40px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); }
        h1 { color: #003366; border-bottom: 3px solid #003366; padding-bottom: 10px; }
        h2 { color: #003366; margin-top: 30px; }
        .summary-box { display: inline-block; padding: 15px 25px; margin: 5px; border-radius: 8px; color: white; text-align: center; min-width: 80px; }
        .pass { background: #228B22; }
        .warn { background: #FFA500; }
        .fail { background: #DC143C; }
        .info { background: #4682B4; }
        .count { font-size: 24px; font-weight: bold; }
        .label { font-size: 12px; }
        .finding { background: #f8f8fa; padding: 15px; margin: 10px 0; border-radius: 5px; border-left: 4px solid #ccc; }
        .finding.status-FAIL { border-left-color: #DC143C; }
        .finding.status-WARN { border-left-color: #FFA500; }
        .finding.status-PASS { border-left-color: #228B22; }
        .finding.status-INFO { border-left-color: #4682B4; }
        .finding-title { font-weight: bold; margin-bottom: 5px; }
        .finding-desc { color: #555; margin-bottom: 5px; }
        .finding-meta { font-size: 11px; color: #888; }
        .finding-impact { color: #6a4f2e; font-style: italic; margin-top: 5px; padding: 6px 10px; background: #fef9f0; border-radius: 3px; }
        .recommendation { background: #fffaef; padding: 10px; margin-top: 10px; border-radius: 3px; font-style: italic; }
        .remediation { background: #f0f4f8; padding: 12px; margin-top: 8px; border-radius: 5px; border: 1px solid #d0d7de; }
        .remediation-header { display: flex; align-items: center; gap: 8px; margin-bottom: 8px; }
        .safety-badge { padding: 2px 8px; border-radius: 3px; font-size: 11px; font-weight: bold; color: white; }
        .safety-safe-apply { background: #228B22; }
        .safety-requires-review { background: #FFA500; }
        .safety-destructive { background: #DC143C; }
        .remediation-commands { list-style: none; padding: 0; margin: 8px 0 0 0; }
        .remediation-commands li { background: #1e1e2e; color: #cdd6f4; padding: 8px 12px; margin: 4px 0; border-radius: 4px; font-family: 'Courier New', monospace; font-size: 12px; }
        .remediation-commands li.confirm { border-left: 3px solid #DC143C; }
        .remediation-cmd-desc { color: #a6adc8; font-size: 11px; margin-bottom: 2px; font-family: 'Segoe UI', Arial, sans-serif; }
        .remediation-prereqs { font-size: 12px; color: #555; margin-top: 6px; }
        .remediation-link { font-size: 12px; margin-top: 6px; }
        .info-table { width: 100%; border-collapse: collapse; }
        .info-table td { padding: 8px; border-bottom: 1px solid #eee; }
        .info-table td:first-child { font-weight: bold; width: 200px; }
        .score-bar { background: #ddd; height: 30px; border-radius: 15px; overflow: hidden; margin: 10px 0; }
        .score-fill { height: 100%; display: flex; align-items: center; justify-content: center; color: white; font-weight: bold; }
        .delta-section { background: #f8f9fa; border: 1px solid #e1e4e8; border-radius: 8px; padding: 15px; margin: 15px 0; }
        .delta-box { display: inline-block; padding: 8px 16px; margin: 4px; border-radius: 6px; border-left: 4px solid; background: #fff; }
        .delta-box.new { border-left-color: #DC143C; }
        .delta-box.resolved { border-left-color: #228B22; }
        .delta-box.regression { border-left-color: #FFA500; }
        .delta-box.improved { border-left-color: #4682B4; }
        .delta-count { font-size: 18px; font-weight: bold; }
        .delta-label { font-size: 11px; color: #666; }
    </style>
</head>
<body>
<div class="container">
`)

	// Title
	buf.WriteString(fmt.Sprintf(`<h1>OpenShift Cluster Assessment Report</h1>
<p style="color: #888;">Generated: %s</p>
`, time.Now().Format("January 2, 2006 at 15:04 MST")))

	// Cluster Info
	info := assessment.Status.ClusterInfo
	buf.WriteString(`<h2>Cluster Information</h2>
<table class="info-table">`)
	buf.WriteString(fmt.Sprintf(`<tr><td>Cluster ID</td><td>%s</td></tr>`, html.EscapeString(info.ClusterID)))
	buf.WriteString(fmt.Sprintf(`<tr><td>OpenShift Version</td><td>%s</td></tr>`, html.EscapeString(info.ClusterVersion)))
	buf.WriteString(fmt.Sprintf(`<tr><td>Platform</td><td>%s</td></tr>`, html.EscapeString(info.Platform)))
	buf.WriteString(fmt.Sprintf(`<tr><td>Update Channel</td><td>%s</td></tr>`, html.EscapeString(info.Channel)))
	buf.WriteString(fmt.Sprintf(`<tr><td>Total Nodes</td><td>%d</td></tr>`, info.NodeCount))
	buf.WriteString(fmt.Sprintf(`<tr><td>Control Plane Nodes</td><td>%d</td></tr>`, info.ControlPlaneNodes))
	buf.WriteString(fmt.Sprintf(`<tr><td>Worker Nodes</td><td>%d</td></tr>`, info.WorkerNodes))
	profileUsed := assessment.Status.Summary.ProfileUsed
	if profileUsed == "" {
		profileUsed = assessment.Spec.Profile
	}
	buf.WriteString(fmt.Sprintf(`<tr><td>Assessment Profile</td><td>%s</td></tr>`, html.EscapeString(profileUsed)))
	buf.WriteString(`</table>`)

	// Summary
	summary := assessment.Status.Summary
	buf.WriteString(`<h2>Assessment Summary</h2>
<div style="margin: 20px 0;">`)
	buf.WriteString(fmt.Sprintf(`<div class="summary-box pass"><div class="count">%d</div><div class="label">PASS</div></div>`, summary.PassCount))
	buf.WriteString(fmt.Sprintf(`<div class="summary-box warn"><div class="count">%d</div><div class="label">WARN</div></div>`, summary.WarnCount))
	buf.WriteString(fmt.Sprintf(`<div class="summary-box fail"><div class="count">%d</div><div class="label">FAIL</div></div>`, summary.FailCount))
	buf.WriteString(fmt.Sprintf(`<div class="summary-box info"><div class="count">%d</div><div class="label">INFO</div></div>`, summary.InfoCount))
	buf.WriteString(`</div>`)
	buf.WriteString(fmt.Sprintf(`<p>Total Checks: %d</p>`, summary.TotalChecks))

	// Score bar
	if summary.Score != nil {
		scoreColor := "#228B22"
		if *summary.Score < 60 {
			scoreColor = "#DC143C"
		} else if *summary.Score < 80 {
			scoreColor = "#FFA500"
		}
		buf.WriteString(fmt.Sprintf(`<div class="score-bar"><div class="score-fill" style="width: %d%%; background: %s;">%d%%</div></div>`, *summary.Score, scoreColor, *summary.Score))
	}

	// Delta section in HTML
	if assessment.Status.Delta != nil {
		delta := assessment.Status.Delta
		buf.WriteString(`<h2>Changes Since Last Run</h2><div class="delta-section">`)
		if delta.ScoreDelta != nil && *delta.ScoreDelta != 0 {
			if *delta.ScoreDelta > 0 {
				buf.WriteString(fmt.Sprintf(`<p style="color: #228B22; font-weight: bold;">Score: +%d points (improved)</p>`, *delta.ScoreDelta))
			} else {
				buf.WriteString(fmt.Sprintf(`<p style="color: #DC143C; font-weight: bold;">Score: %d points (regressed)</p>`, *delta.ScoreDelta))
			}
		}
		buf.WriteString(fmt.Sprintf(`<div class="delta-box new"><div class="delta-count">%d</div><div class="delta-label">New Issues</div></div>`, len(delta.NewFindings)))
		buf.WriteString(fmt.Sprintf(`<div class="delta-box resolved"><div class="delta-count">%d</div><div class="delta-label">Resolved</div></div>`, len(delta.ResolvedFindings)))
		buf.WriteString(fmt.Sprintf(`<div class="delta-box regression"><div class="delta-count">%d</div><div class="delta-label">Regressions</div></div>`, len(delta.RegressionFindings)))
		buf.WriteString(fmt.Sprintf(`<div class="delta-box improved"><div class="delta-count">%d</div><div class="delta-label">Improved</div></div>`, len(delta.ImprovedFindings)))
		buf.WriteString(`</div>`)
	}

	// Detailed Findings
	buf.WriteString(`<h2>Detailed Findings</h2>`)

	statusOrder := []assessmentv1alpha1.FindingStatus{
		assessmentv1alpha1.FindingStatusFail,
		assessmentv1alpha1.FindingStatusWarn,
		assessmentv1alpha1.FindingStatusInfo,
		assessmentv1alpha1.FindingStatusPass,
	}

	// Group findings by status
	findingsByStatus := make(map[assessmentv1alpha1.FindingStatus][]assessmentv1alpha1.Finding)
	for _, f := range assessment.Status.Findings {
		findingsByStatus[f.Status] = append(findingsByStatus[f.Status], f)
	}

	for _, status := range statusOrder {
		for _, f := range findingsByStatus[status] {
			buf.WriteString(fmt.Sprintf(`<div class="finding status-%s">`, f.Status))
			buf.WriteString(fmt.Sprintf(`<div class="finding-title">[%s] %s</div>`, f.Status, html.EscapeString(f.Title)))
			buf.WriteString(fmt.Sprintf(`<div class="finding-desc">%s</div>`, html.EscapeString(f.Description)))

			// Resource/Namespace
			if f.Resource != "" {
				resourceStr := f.Resource
				if f.Namespace != "" {
					resourceStr += " (ns: " + f.Namespace + ")"
				}
				buf.WriteString(fmt.Sprintf(`<div class="finding-meta">Resource: %s</div>`, html.EscapeString(resourceStr)))
			}

			buf.WriteString(fmt.Sprintf(`<div class="finding-meta">Category: %s | Validator: %s</div>`, html.EscapeString(f.Category), html.EscapeString(f.Validator)))

			// Impact
			if f.Impact != "" {
				buf.WriteString(fmt.Sprintf(`<div class="finding-impact">Impact: %s</div>`, html.EscapeString(f.Impact)))
			}

			if f.Recommendation != "" && (f.Status == assessmentv1alpha1.FindingStatusFail || f.Status == assessmentv1alpha1.FindingStatusWarn) {
				buf.WriteString(fmt.Sprintf(`<div class="recommendation">💡 %s</div>`, html.EscapeString(f.Recommendation)))
			}
			if len(f.References) > 0 {
				buf.WriteString(`<div class="finding-meta" style="margin-top: 5px;">References: `)
				for i, ref := range f.References {
					if i > 0 {
						buf.WriteString(", ")
					}
					// Only allow http and https schemes for links to prevent XSS (e.g., javascript:)
					lowerRef := strings.ToLower(ref)
					if strings.HasPrefix(lowerRef, "http://") || strings.HasPrefix(lowerRef, "https://") {
						buf.WriteString(fmt.Sprintf(`<a href="%s">%s</a>`, html.EscapeString(ref), html.EscapeString(truncateURL(ref))))
					} else {
						// Render unsafe URLs as plain text
						buf.WriteString(html.EscapeString(ref))
					}
				}
				buf.WriteString(`</div>`)
			}
			if f.Remediation != nil {
				buf.WriteString(`<div class="remediation">`)
				buf.WriteString(`<div class="remediation-header">`)
				buf.WriteString(`<strong>Remediation</strong>`)
				safetyClass := "safety-" + strings.ReplaceAll(string(f.Remediation.Safety), " ", "-")
				buf.WriteString(fmt.Sprintf(`<span class="safety-badge %s">%s</span>`, html.EscapeString(safetyClass), html.EscapeString(string(f.Remediation.Safety))))
				buf.WriteString(`</div>`)
				if f.Remediation.EstimatedImpact != "" {
					buf.WriteString(fmt.Sprintf(`<div style="font-size: 12px; color: #555; margin-bottom: 6px;">Impact: %s</div>`, html.EscapeString(f.Remediation.EstimatedImpact)))
				}
				if len(f.Remediation.Prerequisites) > 0 {
					buf.WriteString(`<div class="remediation-prereqs"><strong>Prerequisites:</strong><ul>`)
					for _, prereq := range f.Remediation.Prerequisites {
						buf.WriteString(fmt.Sprintf(`<li>%s</li>`, html.EscapeString(prereq)))
					}
					buf.WriteString(`</ul></div>`)
				}
				if len(f.Remediation.Commands) > 0 {
					buf.WriteString(`<ul class="remediation-commands">`)
					for _, cmd := range f.Remediation.Commands {
						liClass := ""
						if cmd.RequiresConfirmation {
							liClass = ` class="confirm"`
						}
						buf.WriteString(fmt.Sprintf(`<li%s>`, liClass))
						if cmd.Description != "" {
							buf.WriteString(fmt.Sprintf(`<div class="remediation-cmd-desc">%s</div>`, html.EscapeString(cmd.Description)))
						}
						if cmd.RequiresConfirmation {
							buf.WriteString("⚠ ")
						}
						buf.WriteString(html.EscapeString(cmd.Command))
						buf.WriteString(`</li>`)
					}
					buf.WriteString(`</ul>`)
				}
				if f.Remediation.DocumentationURL != "" {
					lowerURL := strings.ToLower(f.Remediation.DocumentationURL)
					if strings.HasPrefix(lowerURL, "http://") || strings.HasPrefix(lowerURL, "https://") {
						buf.WriteString(fmt.Sprintf(`<div class="remediation-link"><a href="%s">📖 Documentation</a></div>`, html.EscapeString(f.Remediation.DocumentationURL)))
					}
				}
				buf.WriteString(`</div>`)
			}
			buf.WriteString(`</div>`)
		}
	}

	buf.WriteString(`</div></body></html>`)

	return buf.Bytes(), nil
}

func truncateURL(url string) string {
	if len(url) > 50 {
		return url[:47] + "..."
	}
	return url
}
