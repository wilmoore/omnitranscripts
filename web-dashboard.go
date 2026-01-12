//go:build ignore

package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var startTime = time.Now()
var fileModTime time.Time
var fileMutex sync.RWMutex

// Demo transcript data for testing
var demoTranscripts = map[string]Job{
	"job_1234567890": {
		Transcript: "Welcome to this demonstration video. Today we're going to explore the amazing world of video transcription and how it can transform your content workflow. This technology has revolutionized the way we process and understand multimedia content, making it accessible to everyone.",
		Segments: []TranscriptSegment{
			{Start: 0.0, End: 4.5, Text: "Welcome to this demonstration video."},
			{Start: 4.5, End: 12.8, Text: "Today we're going to explore the amazing world of video transcription"},
			{Start: 12.8, End: 18.2, Text: "and how it can transform your content workflow."},
			{Start: 18.2, End: 25.4, Text: "This technology has revolutionized the way we process"},
			{Start: 25.4, End: 32.1, Text: "and understand multimedia content, making it accessible to everyone."},
		},
	},
}

type Job struct {
	ID         string    `json:"id"`
	VideoID    string    `json:"video_id"`
	URL        string    `json:"url"`
	Title      string    `json:"title"`
	Status     string    `json:"status"`
	Progress   int       `json:"progress"`
	StartTime  time.Time `json:"start_time"`
	UpdateTime time.Time `json:"update_time"`
	LogFile    string    `json:"log_file"`
	OutputDir  string    `json:"output_dir"`
	Duration   string    `json:"duration"`
	FileCount  int       `json:"file_count"`
	FileSize   string    `json:"file_size"`
	Stage      string    `json:"stage"`
	StageProgress map[string]int `json:"stage_progress"`
	CategoryClass string `json:"category_class"`
	CategoryIcon  string `json:"category_icon"`
	StatusText    string `json:"status_text"`
	// Transcript fields
	Transcript    string          `json:"transcript,omitempty"`
	Segments      []TranscriptSegment `json:"segments,omitempty"`
}

type TranscriptSegment struct {
	Start float64 `json:"start"`
	End   float64 `json:"end"`
	Text  string  `json:"text"`
}

const dashboardHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>OmniTranscripts Dashboard</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
            color: #333;
            margin: 0;
            padding: 0;
        }

        .dashboard {
            display: flex;
            min-height: 100vh;
            gap: 0;
        }

        /* Sidebar */
        .sidebar {
            width: 280px;
            background: #2a2d3e;
            color: white;
            padding: 32px 0;
            position: fixed;
            height: 100vh;
            overflow-y: auto;
            z-index: 1000;
            left: 0;
            top: 0;
        }

        .brand-section {
            padding: 0 32px;
            margin-bottom: 48px;
            position: relative;
        }

        .brand-logo {
            text-align: center;
        }

        .logo-icon {
            font-size: 48px;
            margin-bottom: 16px;
        }

        .brand-name {
            font-size: 28px;
            font-weight: 700;
            margin-bottom: 8px;
            color: white;
            letter-spacing: -0.5px;
        }

        .brand-tagline {
            font-size: 14px;
            color: #9ca3af;
            font-weight: 400;
        }

        .nav-menu {
            padding: 0 32px;
        }

        .nav-item {
            padding: 14px 0;
            font-size: 16px;
            color: #9ca3af;
            cursor: pointer;
            transition: color 0.2s;
            border-bottom: 1px solid rgba(156, 163, 175, 0.1);
        }

        .nav-item:last-child {
            border-bottom: none;
        }

        .nav-item:hover {
            color: white;
        }

        .nav-item.active {
            color: white;
            font-weight: 600;
        }

        /* Main Content */
        .main-content {
            flex: 1;
            margin-left: 280px;
            margin-right: 380px;
            padding: 0;
            background: #f5f7fa;
            min-height: 100vh;
            position: relative;
        }

        .content-header {
            display: flex;
            justify-content: space-between;
            align-items: flex-start;
            margin-bottom: 32px;
            padding: 32px 32px 24px 32px;
            border-bottom: 1px solid #f1f5f9;
        }

        .page-title {
            font-size: 36px;
            font-weight: 600;
            color: #2a2d3e;
            margin-bottom: 4px;
            line-height: 1.2;
        }

        .page-subtitle {
            font-size: 16px;
            color: #64748b;
            font-weight: 400;
        }

        .header-info {
            flex: 1;
        }

        .status-indicator {
            display: flex;
            align-items: center;
            gap: 10px;
            background: #f0f9ff;
            padding: 12px 16px;
            border-radius: 12px;
            border: 1px solid #bae6fd;
            min-width: 120px;
        }

        .status-dot {
            width: 8px;
            height: 8px;
            background: #10b981;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0%, 100% { opacity: 1; }
            50% { opacity: 0.5; }
        }

        .status-text {
            font-size: 14px;
            color: #0369a1;
            font-weight: 500;
        }

        /* Chart Area */
        .chart-container {
            margin: 0 32px 24px 32px;
            padding: 20px;
            background: #f8fafc;
            border-radius: 16px;
            border: 1px solid #e2e8f0;
        }

        .chart {
            display: flex;
            align-items: end;
            gap: 6px;
            height: 120px;
            margin-bottom: 16px;
            padding: 0 8px;
        }

        .chart-bar {
            background: #3b82f6;
            border-radius: 6px 6px 0 0;
            min-width: 14px;
            transition: all 0.2s;
            opacity: 0.8;
        }

        .chart-bar:hover {
            background: #2563eb;
            opacity: 1;
            transform: translateY(-2px);
        }

        /* Transaction Lists */
        .transactions {
            margin: 0 32px 32px 32px;
        }

        .date-header {
            font-size: 20px;
            font-weight: 600;
            color: #1e293b;
            margin-bottom: 24px;
            padding-bottom: 16px;
            border-bottom: 1px solid #e2e8f0;
            position: relative;
        }

        .date-header::after {
            content: '⋯';
            position: absolute;
            right: 0;
            top: 0;
            color: #64748b;
            font-size: 20px;
        }

        .transaction-item {
            display: flex;
            align-items: center;
            justify-content: space-between;
            padding: 16px 0;
            border-bottom: 1px solid #f1f5f9;
            transition: background 0.2s;
        }

        .transaction-item:hover {
            background: #f8fafc;
            margin: 0 -16px;
            padding: 16px;
            border-radius: 12px;
            cursor: pointer;
        }

        .transaction-item:last-child {
            border-bottom: none;
        }

        .clickable {
            cursor: pointer;
            transition: color 0.2s;
        }

        .clickable:hover {
            color: #3b82f6;
        }

        .video-id {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            font-size: 12px;
            color: #6b7280;
            background: #f3f4f6;
            padding: 4px 8px;
            border-radius: 6px;
            border: 1px solid #e5e7eb;
            cursor: pointer;
            transition: all 0.2s;
            display: inline-flex;
            align-items: center;
            gap: 6px;
            margin-left: 8px;
        }

        .video-id:hover {
            background: #e5e7eb;
            border-color: #d1d5db;
        }

        .copy-icon {
            font-size: 10px;
            opacity: 0.6;
        }

        .transaction-meta {
            display: flex;
            align-items: center;
            gap: 12px;
            flex-shrink: 0;
        }

        .transaction-left {
            display: flex;
            align-items: center;
            flex: 1;
            min-width: 0;
        }

        .transaction-icon {
            width: 48px;
            height: 48px;
            border-radius: 50%;
            margin-right: 20px;
            display: flex;
            align-items: center;
            justify-content: center;
            font-size: 20px;
            color: white;
            flex-shrink: 0;
        }

        .transaction-icon.grocery { background: #3b82f6; }
        .transaction-icon.transport { background: #8b5cf6; }
        .transaction-icon.housing { background: #f97316; }
        .transaction-icon.food { background: #ef4444; }
        .transaction-icon.entertainment { background: #10b981; }

        .transaction-details {
            flex: 1;
            min-width: 0;
        }

        .transaction-title {
            font-size: 16px;
            font-weight: 600;
            color: #1e293b;
            margin-bottom: 4px;
            line-height: 1.3;
        }

        .transaction-subtitle {
            font-size: 14px;
            color: #64748b;
            line-height: 1.4;
        }

        .transaction-amount {
            font-size: 16px;
            font-weight: 600;
            color: #1e293b;
            flex-shrink: 0;
            margin-left: 16px;
        }

        /* Right Sidebar */
        .right-sidebar {
            width: 380px;
            background: #f8fafc;
            padding: 52px 32px 32px 32px;
            position: fixed;
            right: 0;
            top: 0;
            height: 100vh;
            overflow-y: auto;
            z-index: 999;
            border-left: 1px solid #e2e8f0;
        }

        .stats-section {
            margin-bottom: 48px;
        }

        .stats-title {
            font-size: 18px;
            font-weight: 600;
            color: #1e293b;
            margin-bottom: 24px;
            padding-bottom: 12px;
            border-bottom: 1px solid #e2e8f0;
        }

        .stat-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 20px;
            padding: 4px 0;
        }

        .stat-label {
            font-size: 14px;
            color: #64748b;
            font-weight: 500;
        }

        .stat-value {
            font-size: 15px;
            font-weight: 600;
            color: #1e293b;
        }

        .stat-bar {
            width: 100%;
            height: 6px;
            background: #e2e8f0;
            border-radius: 3px;
            margin-top: 10px;
            overflow: hidden;
        }

        .stat-progress {
            height: 100%;
            border-radius: 3px;
            transition: width 0.3s ease;
        }

        .stat-progress.completed { background: #10b981; }
        .stat-progress.running { background: #f59e0b; }
        .stat-progress.failed { background: #ef4444; }
        .stat-progress.queued { background: #64748b; }

        .tips-section {
            background: white;
            border-radius: 16px;
            padding: 30px;
            position: relative;
            overflow: hidden;
        }

        .tips-title {
            font-size: 18px;
            font-weight: 600;
            color: #2a2d3e;
            margin-bottom: 12px;
        }

        .tips-subtitle {
            font-size: 14px;
            color: #6b7280;
            line-height: 1.5;
            margin-bottom: 24px;
        }

        .tips-btn {
            background: #2a2d3e;
            color: white;
            border: none;
            padding: 12px 40px;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 600;
            cursor: pointer;
            transition: background 0.2s;
        }

        .tips-btn:hover {
            background: #1f2937;
        }

        .tips-illustration {
            position: absolute;
            right: -20px;
            top: -20px;
            opacity: 0.1;
            font-size: 80px;
        }

        /* New Metrics Styles */
        .metric-grid {
            display: grid;
            grid-template-columns: 1fr 1fr;
            gap: 16px;
        }

        .metric-card {
            background: white;
            border-radius: 16px;
            padding: 24px 20px;
            text-align: center;
            border: 1px solid #e2e8f0;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
        }

        .metric-value {
            font-size: 28px;
            font-weight: 700;
            color: #1e293b;
            margin-bottom: 8px;
            line-height: 1;
        }

        .metric-label {
            font-size: 12px;
            color: #64748b;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.5px;
        }

        .usage-stats {
            background: white;
            border-radius: 16px;
            padding: 24px;
            border: 1px solid #e2e8f0;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
        }

        .usage-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 12px 0;
            border-bottom: 1px solid #f1f5f9;
        }

        .usage-item:last-child {
            border-bottom: none;
            padding-bottom: 0;
        }

        .usage-item:first-child {
            padding-top: 0;
        }

        .usage-label {
            font-size: 14px;
            color: #64748b;
            font-weight: 500;
        }

        .usage-value {
            font-size: 14px;
            font-weight: 600;
            color: #1e293b;
        }

        .health-metrics {
            margin-bottom: 20px;
        }

        .health-item {
            display: flex;
            justify-content: space-between;
            align-items: center;
            margin-bottom: 12px;
        }

        .health-label {
            font-size: 14px;
            color: #6b7280;
        }

        .health-value {
            font-size: 14px;
            font-weight: 600;
            color: #2a2d3e;
        }

        .health-value.healthy {
            color: #10b981;
        }

        .health-value.warning {
            color: #f59e0b;
        }

        .health-value.critical {
            color: #ef4444;
        }

        /* Add Job Card */
        .add-job-card {
            margin: 0 32px 32px 32px;
            background: white;
            border-radius: 16px;
            border: 1px solid #e2e8f0;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
            overflow: hidden;
        }

        .add-job-header {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 24px 32px 20px 32px;
            border-bottom: 1px solid #f1f5f9;
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            color: white;
        }

        .add-job-title {
            font-size: 18px;
            font-weight: 600;
            margin: 0;
            color: white;
        }

        .add-job-content {
            padding: 24px 32px;
        }

        .input-group {
            display: flex;
            gap: 16px;
            align-items: stretch;
            margin-bottom: 20px;
        }

        .url-input {
            flex: 1;
            padding: 14px 18px;
            border: 2px solid #e2e8f0;
            border-radius: 12px;
            background: white;
            color: #374151;
            font-size: 15px;
            transition: all 0.2s;
            font-family: inherit;
        }

        .url-input:focus {
            outline: none;
            border-color: #667eea;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .url-input::placeholder {
            color: #9ca3af;
        }

        .btn-primary {
            background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
            border: none;
            color: white;
            padding: 14px 24px;
            border-radius: 12px;
            font-size: 15px;
            font-weight: 600;
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 8px;
            white-space: nowrap;
            box-shadow: 0 2px 4px rgba(102, 126, 234, 0.2);
        }

        .btn-primary:hover {
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(102, 126, 234, 0.3);
        }

        .btn-icon {
            font-size: 18px;
            font-weight: 400;
        }

        .add-job-info {
            display: flex;
            gap: 24px;
            flex-wrap: wrap;
        }

        .info-item {
            display: flex;
            align-items: center;
            gap: 8px;
            padding: 12px 16px;
            background: #f8fafc;
            border-radius: 10px;
            border: 1px solid #e2e8f0;
        }

        .info-icon {
            font-size: 16px;
        }

        .info-text {
            font-size: 13px;
            color: #64748b;
            font-weight: 500;
        }

        /* Transcriptions View */
        .transcriptions-header {
            margin: 0 32px 32px 32px;
        }

        .search-filter-bar {
            display: flex;
            justify-content: space-between;
            align-items: center;
            gap: 24px;
            padding: 20px 24px;
            background: white;
            border-radius: 16px;
            border: 1px solid #e2e8f0;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
        }

        .search-box {
            position: relative;
            flex: 1;
            max-width: 400px;
        }

        .search-input {
            width: 100%;
            padding: 12px 16px 12px 44px;
            border: 2px solid #e2e8f0;
            border-radius: 12px;
            background: #f8fafc;
            color: #374151;
            font-size: 14px;
            transition: all 0.2s;
        }

        .search-input:focus {
            outline: none;
            border-color: #667eea;
            background: white;
            box-shadow: 0 0 0 3px rgba(102, 126, 234, 0.1);
        }

        .search-icon {
            position: absolute;
            left: 16px;
            top: 50%;
            transform: translateY(-50%);
            font-size: 16px;
            color: #9ca3af;
        }

        .filter-buttons {
            display: flex;
            gap: 8px;
        }

        .filter-btn {
            padding: 8px 16px;
            border: 1px solid #e2e8f0;
            border-radius: 8px;
            background: white;
            color: #64748b;
            font-size: 14px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
        }

        .filter-btn:hover {
            background: #f8fafc;
            border-color: #d1d5db;
        }

        .filter-btn.active {
            background: #667eea;
            border-color: #667eea;
            color: white;
        }

        .transcriptions-grid {
            margin: 0 32px;
            display: grid;
            grid-template-columns: repeat(auto-fill, minmax(380px, 1fr));
            gap: 24px;
        }

        .transcription-card {
            background: white;
            border-radius: 16px;
            border: 1px solid #e2e8f0;
            overflow: hidden;
            transition: all 0.2s;
            box-shadow: 0 1px 3px rgba(0, 0, 0, 0.05);
        }

        .transcription-card:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(0, 0, 0, 0.1);
            border-color: #d1d5db;
        }

        .transcription-header {
            padding: 20px 24px;
        }

        .transcription-info {
            margin-bottom: 16px;
        }

        .transcription-title {
            font-size: 16px;
            font-weight: 600;
            color: #1e293b;
            margin-bottom: 12px;
            line-height: 1.4;
            display: -webkit-box;
            -webkit-line-clamp: 2;
            -webkit-box-orient: vertical;
            overflow: hidden;
        }

        .transcription-meta {
            display: flex;
            flex-wrap: wrap;
            gap: 16px;
            margin-bottom: 16px;
        }

        .meta-item {
            display: flex;
            align-items: center;
            gap: 6px;
            font-size: 12px;
            color: #64748b;
        }

        .meta-icon {
            font-size: 14px;
        }

        .transcription-actions {
            display: flex;
            gap: 12px;
        }

        .action-btn {
            padding: 8px 16px;
            border: none;
            border-radius: 8px;
            font-size: 13px;
            font-weight: 500;
            cursor: pointer;
            transition: all 0.2s;
            display: flex;
            align-items: center;
            gap: 6px;
        }

        .view-btn {
            background: #f0f9ff;
            color: #0369a1;
            border: 1px solid #bae6fd;
        }

        .view-btn:hover {
            background: #e0f2fe;
            border-color: #7dd3fc;
        }

        .download-btn {
            background: #f0fdf4;
            color: #15803d;
            border: 1px solid #bbf7d0;
        }

        .download-btn:hover {
            background: #dcfce7;
            border-color: #86efac;
        }

        .video-id-badge {
            background: #f8fafc;
            border-top: 1px solid #f1f5f9;
            padding: 12px 24px;
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            font-size: 12px;
            color: #64748b;
            cursor: pointer;
            transition: background 0.2s;
        }

        .video-id-badge:hover {
            background: #f1f5f9;
        }

        .empty-state {
            text-align: center;
            padding: 64px 32px;
            color: #64748b;
            font-size: 16px;
            grid-column: 1 / -1;
        }

        .btn {
            padding: 12px 24px;
            border: none;
            border-radius: 10px;
            background: #3b82f6;
            color: white;
            cursor: pointer;
            font-size: 14px;
            font-weight: 500;
            transition: all 0.2s;
            white-space: nowrap;
        }

        .btn:hover {
            background: #2563eb;
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(59, 130, 246, 0.25);
        }

        .btn-secondary {
            background: #64748b;
        }

        .btn-secondary:hover {
            background: #475569;
            transform: translateY(-1px);
            box-shadow: 0 4px 12px rgba(100, 116, 139, 0.25);
        }

        input[type="text"] {
            padding: 12px 16px;
            border: 1px solid #d1d5db;
            border-radius: 10px;
            background: white;
            color: #374151;
            flex: 1;
            max-width: 400px;
            font-size: 14px;
            transition: border-color 0.2s, box-shadow 0.2s;
        }

        input[type="text"]:focus {
            outline: none;
            border-color: #3b82f6;
            box-shadow: 0 0 0 3px rgba(59, 130, 246, 0.1);
        }

        .live-indicator {
            margin-left: auto;
            color: #10b981;
            font-size: 14px;
            font-weight: 500;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .live-dot {
            width: 8px;
            height: 8px;
            background: #10b981;
            border-radius: 50%;
            animation: pulse 2s infinite;
        }

        @keyframes pulse {
            0% { opacity: 1; }
            50% { opacity: 0.5; }
            100% { opacity: 1; }
        }

        /* Mobile menu toggle */
        .mobile-menu-toggle {
            display: none;
            position: fixed;
            top: 20px;
            left: 20px;
            z-index: 1001;
            background: #2a2d3e;
            color: white;
            border: none;
            padding: 12px;
            border-radius: 8px;
            cursor: pointer;
        }

        .sidebar-overlay {
            display: none;
            position: fixed;
            top: 0;
            left: 0;
            width: 100%;
            height: 100%;
            background: rgba(0, 0, 0, 0.5);
            z-index: 999;
        }

        /* Responsive */
        @media (max-width: 1200px) {
            .right-sidebar {
                display: none;
            }
            .main-content {
                margin-right: 40px;
            }
        }

        @media (max-width: 768px) {
            .mobile-menu-toggle {
                display: block;
            }

            .sidebar {
                transform: translateX(-100%);
                transition: transform 0.3s ease;
            }

            .sidebar.open {
                transform: translateX(0);
            }

            .sidebar-overlay.open {
                display: block;
            }

            .main-content {
                margin-left: 0;
                padding: 80px 20px 20px 20px;
                margin-top: 0;
                border-radius: 0;
            }

            .content-header {
                flex-direction: column;
                gap: 20px;
            }

            .page-title {
                font-size: 32px;
            }

            .chart {
                gap: 4px;
            }

            .chart-bar {
                min-width: 12px;
            }

            .transaction-item {
                padding: 15px 0;
            }

            .transaction-icon {
                width: 40px;
                height: 40px;
                margin-right: 15px;
                font-size: 16px;
            }

            .transaction-title {
                font-size: 16px;
            }

            .transaction-amount {
                font-size: 16px;
            }

            .controls {
                flex-direction: column;
                gap: 15px;
                align-items: stretch;
            }

            input[type="text"] {
                max-width: none;
            }

            .add-job-card {
                margin: 0 20px 24px 20px;
            }

            .add-job-header {
                padding: 20px 24px 16px 24px;
            }

            .add-job-content {
                padding: 20px 24px;
            }

            .input-group {
                flex-direction: column;
                gap: 12px;
            }

            .add-job-info {
                flex-direction: column;
                gap: 12px;
            }

            .transcriptions-header {
                margin: 0 20px 24px 20px;
            }

            .search-filter-bar {
                flex-direction: column;
                gap: 16px;
                padding: 16px 20px;
            }

            .search-box {
                max-width: none;
            }

            .filter-buttons {
                flex-wrap: wrap;
                justify-content: center;
            }

            .transcriptions-grid {
                margin: 0 20px;
                grid-template-columns: 1fr;
                gap: 16px;
            }

            .transcription-card {
                margin: 0;
            }
        }

        @media (max-width: 480px) {
            .brand-section {
                padding: 0 20px;
            }

            .nav-menu {
                padding: 0 20px;
            }

            .main-content {
                padding: 80px 15px 15px 15px;
            }

            .page-title {
                font-size: 28px;
            }

            .sidebar {
                width: 260px;
            }

            .brand-name {
                font-size: 24px;
            }

            .logo-icon {
                font-size: 40px;
            }
        }

        /* Business Metrics Styles */
        .chart-header h3 {
            font-size: 18px;
            font-weight: 600;
            color: #1f2937;
            margin-bottom: 4px;
        }

        .chart-subtitle {
            font-size: 14px;
            color: #6b7280;
            margin-bottom: 16px;
        }

        .business-metrics {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(180px, 1fr));
            gap: 16px;
            margin-bottom: 20px;
        }

        .metric-card {
            background: white;
            padding: 16px;
            border-radius: 12px;
            border: 1px solid #e5e7eb;
            text-align: center;
        }

        .metric-value {
            font-size: 24px;
            font-weight: 700;
            color: #1f2937;
            margin-bottom: 4px;
        }

        .metric-label {
            font-size: 13px;
            color: #6b7280;
            margin-bottom: 6px;
        }

        .metric-change {
            font-size: 12px;
            font-weight: 500;
        }

        .metric-change.positive {
            color: #10b981;
        }

        .metric-trend {
            font-size: 12px;
            color: #6b7280;
        }

        .metric-insight {
            font-size: 12px;
            color: #8b5cf6;
            font-weight: 500;
        }

        .metric-status {
            font-size: 12px;
            font-weight: 500;
            padding: 2px 6px;
            border-radius: 4px;
        }

        .metric-status.excellent {
            background: #dcfce7;
            color: #166534;
        }

        .metric-status.good {
            background: #fef3c7;
            color: #92400e;
        }

        .metric-status.needs-attention {
            background: #fee2e2;
            color: #991b1b;
        }
    </style>
</head>
<body>
    <div class="dashboard">
        <!-- Sidebar -->
        <div class="sidebar">
            <div class="brand-section">
                <div class="brand-logo">
                    <div class="logo-icon">🎬</div>
                    <div class="brand-name">OmniTranscripts</div>
                    <div class="brand-tagline">AI-Powered Transcription</div>
                </div>
            </div>

            <nav class="nav-menu">
                <div class="nav-item active" onclick="showDashboard()">📊 Dashboard</div>
                <div class="nav-item" onclick="showTranscriptions()">🎥 Transcriptions</div>
                <div class="nav-item" onclick="showQueue()">📝 Queue</div>
            </nav>
        </div>

        <!-- Main Content -->
        <div class="main-content">
            <div class="content-header">
                <div class="header-info">
                    <h1 class="page-title">Transcription Dashboard</h1>
                    <p class="page-subtitle">Real-time video transcription monitoring</p>
                </div>
                <div class="status-indicator">
                    <div class="status-dot"></div>
                    <span class="status-text">API Online</span>
                </div>
            </div>

            <!-- Dashboard View -->
            <div id="dashboard-view">
                <!-- Add Job Card -->
                <div class="add-job-card">
                    <div class="add-job-header">
                        <h3 class="add-job-title">🎬 Add New Transcription</h3>
                    </div>
                    <div class="add-job-content">
                        <div class="input-group">
                            <input type="text" id="url-input" placeholder="Enter YouTube URL to add new transcription job..." class="url-input">
                            <button class="btn btn-primary" onclick="addJob()">
                                <span class="btn-icon">+</span>
                                Add Job
                            </button>
                        </div>
                    </div>
                </div>

                <!-- Business Insights Chart -->
                <div class="chart-container">
                    <div class="chart-header">
                        <h3>System Performance Metrics</h3>
                        <p class="chart-subtitle">Real-time transcription system monitoring</p>
                    </div>
                    <div class="business-metrics">
                        <div class="metric-card">
                            <div class="metric-value">{{.JobsToday}}</div>
                            <div class="metric-label">Jobs Today</div>
                            <div class="metric-trend">{{if gt .JobsToday .JobsYesterday}}📈 Active{{else}}📊 Steady{{end}}</div>
                        </div>
                        <div class="metric-card">
                            <div class="metric-value">{{printf "%.1f" .SuccessRate}}%</div>
                            <div class="metric-label">Success Rate</div>
                            <div class="metric-status {{if gt .SuccessRate 95.0}}excellent{{else if gt .SuccessRate 85.0}}good{{else}}needs-attention{{end}}">
                                {{if gt .SuccessRate 95.0}}Excellent{{else if gt .SuccessRate 85.0}}Good{{else}}Needs Attention{{end}}
                            </div>
                        </div>
                        <div class="metric-card">
                            <div class="metric-value">{{.AvgResponseTime}}ms</div>
                            <div class="metric-label">Avg Response Time</div>
                            <div class="metric-insight">{{if lt .AvgResponseTime 300}}Fast{{else if lt .AvgResponseTime 500}}Normal{{else}}Slow{{end}}</div>
                        </div>
                        <div class="metric-card">
                            <div class="metric-value">{{.TotalJobs}}</div>
                            <div class="metric-label">Total Processed</div>
                            <div class="metric-trend">📁 All time</div>
                        </div>
                    </div>
                </div>

                <!-- Transcription Jobs -->
                <div class="transactions">
                    <div class="date-header">Recent Transcriptions</div>

                    {{range .Jobs}}
                    <div class="transaction-item" onclick="showJobDetails('{{.ID}}')">
                        <div class="transaction-left">
                            <div class="transaction-icon {{.CategoryClass}}">{{.CategoryIcon}}</div>
                            <div class="transaction-details">
                                <div class="transaction-title clickable">{{.Title}}</div>
                                <div class="transaction-subtitle">{{.UpdateTime.Format "15:04"}} • {{.StatusText}} • {{.Duration}}</div>
                            </div>
                        </div>
                        <div class="transaction-meta">
                            <span class="video-id" onclick="copyToClipboard('{{.VideoID}}', event)" title="Click to copy Video ID">
                                {{.VideoID}} <span class="copy-icon">📋</span>
                            </span>
                            <div class="transaction-amount">{{.Progress}}%</div>
                        </div>
                    </div>
                    {{end}}
                </div>
            </div>

            <!-- Transcriptions View -->
            <div id="transcriptions-view" style="display: none;">
                <div class="transcriptions-header">
                    <div class="search-filter-bar">
                        <div class="search-box">
                            <input type="text" placeholder="Search transcriptions..." class="search-input">
                            <span class="search-icon">🔍</span>
                        </div>
                        <div class="filter-buttons">
                            <button class="filter-btn active" onclick="filterTranscriptions('all')">All</button>
                            <button class="filter-btn" onclick="filterTranscriptions('today')">Today</button>
                            <button class="filter-btn" onclick="filterTranscriptions('week')">This Week</button>
                            <button class="filter-btn" onclick="filterTranscriptions('month')">This Month</button>
                        </div>
                    </div>
                </div>
                <div id="transcriptions-list" class="transcriptions-grid">
                    <!-- Transcription cards will be populated here -->
                </div>
            </div>
        </div>

        <!-- Right Sidebar -->
        <div class="right-sidebar">
            <!-- Transcription Status -->
            <div class="stats-section">
                <h3 class="stats-title">Transcription Status</h3>

                <div class="stat-item">
                    <span class="stat-label">✅ Completed</span>
                    <span class="stat-value" id="completed-stat">{{.CompletedJobs}}</span>
                </div>
                <div class="stat-bar"><div class="stat-progress completed" id="completed-progress" style="width: {{.CompletedPercentage}}%"></div></div>

                <div class="stat-item">
                    <span class="stat-label">⏳ Processing</span>
                    <span class="stat-value" id="running-stat">{{.RunningJobs}}</span>
                </div>
                <div class="stat-bar"><div class="stat-progress running" id="running-progress" style="width: {{.RunningPercentage}}%"></div></div>

                <div class="stat-item">
                    <span class="stat-label">❌ Failed</span>
                    <span class="stat-value" id="failed-stat">{{.FailedJobs}}</span>
                </div>
                <div class="stat-bar"><div class="stat-progress failed" id="failed-progress" style="width: {{.FailedPercentage}}%"></div></div>

                <div class="stat-item">
                    <span class="stat-label">📋 Queued</span>
                    <span class="stat-value" id="queued-stat">{{.QueuedJobs}}</span>
                </div>
                <div class="stat-bar"><div class="stat-progress queued" id="queued-progress" style="width: {{.QueuedPercentage}}%"></div></div>
            </div>

            <!-- RapidAPI Stats -->
            <div class="stats-section">
                <h3 class="stats-title">RapidAPI Metrics</h3>

                <div class="usage-stats">
                    <div class="usage-item">
                        <span class="usage-label">🔥 Today</span>
                        <span class="usage-value">{{.JobsToday}} calls</span>
                    </div>
                    <div class="usage-item">
                        <span class="usage-label">📊 This Week</span>
                        <span class="usage-value">{{.JobsThisWeek}} calls</span>
                    </div>
                    <div class="usage-item">
                        <span class="usage-label">👥 Active Users</span>
                        <span class="usage-value">{{.ActiveUsers}}</span>
                    </div>
                    <div class="usage-item">
                        <span class="usage-label">⚡ Response Time</span>
                        <span class="usage-value">{{.AvgResponseTime}}ms</span>
                    </div>
                </div>
            </div>

            <!-- System Status -->
            <div class="tips-section">
                <div class="tips-illustration">🎬</div>
                <h4 class="tips-title">System Status</h4>
                <div class="health-metrics">
                    <div class="health-item">
                        <span class="health-label">🕐 Uptime</span>
                        <span class="health-value">{{.Uptime}}</span>
                    </div>
                    <div class="health-item">
                        <span class="health-label">📋 Queue Health</span>
                        <span class="health-value {{.QueueHealthClass}}">{{.QueueHealth}}</span>
                    </div>
                </div>
            </div>
        </div>
    </div>

    <script>
        // SSE connection for real-time updates
        let eventSource;
        let reconnectTimeout;

        function connectSSE() {
            if (eventSource) {
                eventSource.close();
            }

            eventSource = new EventSource('/events');

            eventSource.onopen = function() {
                document.getElementById('connection-status').textContent = 'Live Updates';
                document.querySelector('.live-indicator').style.color = '#10b981';
                clearTimeout(reconnectTimeout);
            };

            eventSource.onerror = function() {
                document.getElementById('connection-status').textContent = 'Reconnecting...';
                document.querySelector('.live-indicator').style.color = '#f59e0b';

                // Attempt to reconnect after 3 seconds
                reconnectTimeout = setTimeout(() => {
                    connectSSE();
                }, 3000);
            };

            eventSource.addEventListener('jobs', function(event) {
                const jobs = JSON.parse(event.data);
                updateJobsList(jobs);
            });

            eventSource.addEventListener('stats', function(event) {
                const stats = JSON.parse(event.data);
                updateDashboardStats(stats);

                // Flash indicator to show live update
                flashLiveIndicator();
            });
        }

        function updateJobsList(jobs) {
            // Store jobs globally for transcriptions view
            window.currentJobs = jobs;

            const container = document.querySelector('.transactions');
            if (!container) return;

            // Find the transactions container (skip the header)
            const header = container.querySelector('.date-header');

            // Clear existing transactions (keep header)
            while (container.children.length > 1) {
                container.removeChild(container.lastChild);
            }

            // Add updated transactions
            jobs.forEach(job => {
                const jobElement = createJobElement(job);
                container.appendChild(jobElement);
            });

            // Update transcriptions view if currently active
            if (document.getElementById('transcriptions-view').style.display !== 'none') {
                loadTranscriptionsView();
            }
        }

        function createJobElement(job) {
            const div = document.createElement('div');
            div.className = 'transaction-item';
            div.setAttribute('onclick', 'showJobDetails(\'' + job.id + '\')');

            const statusIconMap = {
                'completed': '✅',
                'failed': '❌',
                'queued': '⏳',
                'running': '🔄'
            };

            const updateTime = new Date(job.update_time);
            const timeStr = updateTime.toLocaleTimeString('en-US', {hour12: false, hour: '2-digit', minute: '2-digit'});

            div.innerHTML =
                '<div class="transaction-left">' +
                    '<div class="transaction-icon ' + job.category_class + '">' + job.category_icon + '</div>' +
                    '<div class="transaction-details">' +
                        '<div class="transaction-title clickable">' + job.title + '</div>' +
                        '<div class="transaction-subtitle">' + timeStr + ' • ' + job.status_text + ' • ' + job.duration + '</div>' +
                    '</div>' +
                '</div>' +
                '<div class="transaction-meta">' +
                    '<span class="video-id" onclick="copyToClipboard(\'' + job.video_id + '\', event)" title="Click to copy Video ID">' +
                        job.video_id + ' <span class="copy-icon">📋</span>' +
                    '</span>' +
                    '<div class="transaction-amount">' + job.progress + '%</div>' +
                '</div>';
            return div;
        }

        function updateDashboardStats(stats) {
            // Update status counts
            updateStatValue('completed-stat', stats.CompletedJobs);
            updateStatValue('running-stat', stats.RunningJobs);
            updateStatValue('failed-stat', stats.FailedJobs);
            updateStatValue('queued-stat', stats.QueuedJobs);

            // Update percentages
            updateStatProgress('completed-progress', stats.CompletedPercentage);
            updateStatProgress('running-progress', stats.RunningPercentage);
            updateStatProgress('failed-progress', stats.FailedPercentage);
            updateStatProgress('queued-progress', stats.QueuedPercentage);

            // Update business metrics in main area
            updateMetricValue('revenue-today', '$' + Math.floor(stats.RevenueToday));
            updateMetricValue('jobs-processed', stats.JobsToday);
            updateMetricValue('avg-revenue-job', '$' + stats.AvgRevenuePerJob.toFixed(2));
            updateMetricValue('success-rate', stats.SuccessRate.toFixed(1) + '%');

            // Update growth trend
            updateMetricChange('revenue-growth', stats.RevenueGrowth);
            updateJobTrend('jobs-trend', stats.JobsToday, stats.JobsYesterday);
            updateRevenueInsight('revenue-insight', stats.AvgRevenuePerJob);
            updateSuccessStatus('success-status', stats.SuccessRate);

            // Update RapidAPI sidebar metrics
            updateUsageValue('jobs-today-sidebar', stats.JobsToday + ' calls');
            updateUsageValue('jobs-week-sidebar', stats.JobsThisWeek + ' calls');
            updateUsageValue('active-users', stats.ActiveUsers);
            updateUsageValue('response-time', stats.AvgResponseTime + 'ms');

            // Update system health
            updateHealthValue('uptime', stats.Uptime);
            updateHealthValue('queue-health', stats.QueueHealth, stats.QueueHealthClass);
        }

        function updateStatValue(id, value) {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        }

        function updateStatProgress(id, percentage) {
            const element = document.getElementById(id);
            if (element) element.style.width = percentage + '%';
        }

        function updateTextContent(id, value) {
            const element = document.getElementById(id);
            if (element) element.textContent = value;
        }

        function updateMetricValue(className, value) {
            const elements = document.querySelectorAll('.metric-card .metric-value');
            elements.forEach(element => {
                const card = element.closest('.metric-card');
                const label = card.querySelector('.metric-label');
                if (label) {
                    if ((className === 'revenue-today' && label.textContent.includes('Revenue Today')) ||
                        (className === 'jobs-processed' && label.textContent.includes('Jobs Processed')) ||
                        (className === 'avg-revenue-job' && label.textContent.includes('Avg Revenue/Job')) ||
                        (className === 'success-rate' && label.textContent.includes('Success Rate'))) {
                        element.textContent = value;
                    }
                }
            });
        }

        function updateMetricChange(className, growthValue) {
            const elements = document.querySelectorAll('.metric-change');
            elements.forEach(element => {
                if (element.textContent.includes('↗')) {
                    element.textContent = '↗ ' + growthValue.toFixed(1) + '%';
                    element.className = growthValue >= 0 ? 'metric-change positive' : 'metric-change';
                }
            });
        }

        function updateJobTrend(className, todayJobs, yesterdayJobs) {
            const elements = document.querySelectorAll('.metric-trend');
            elements.forEach(element => {
                if (element.textContent.includes('Growth') || element.textContent.includes('Decline')) {
                    if (todayJobs > yesterdayJobs) {
                        element.textContent = '📈 Growth vs yesterday';
                    } else {
                        element.textContent = '📉 Decline vs yesterday';
                    }
                }
            });
        }

        function updateRevenueInsight(className, avgRevenue) {
            const elements = document.querySelectorAll('.metric-insight');
            elements.forEach(element => {
                if (element.textContent.includes('tier')) {
                    element.textContent = avgRevenue > 5.0 ? 'Premium tier' : 'Standard tier';
                }
            });
        }

        function updateSuccessStatus(className, successRate) {
            const elements = document.querySelectorAll('.metric-status');
            elements.forEach(element => {
                let status, className;
                if (successRate > 95.0) {
                    status = 'Excellent';
                    className = 'metric-status excellent';
                } else if (successRate > 85.0) {
                    status = 'Good';
                    className = 'metric-status good';
                } else {
                    status = 'Needs Attention';
                    className = 'metric-status needs-attention';
                }
                element.textContent = status;
                element.className = className;
            });
        }

        function updateUsageValue(className, value) {
            const elements = document.querySelectorAll('.usage-value');
            elements.forEach(element => {
                const usageItem = element.closest('.usage-item');
                const label = usageItem.querySelector('.usage-label');
                if (label) {
                    if ((className === 'jobs-today-sidebar' && label.textContent.includes('Today')) ||
                        (className === 'jobs-week-sidebar' && label.textContent.includes('Week')) ||
                        (className === 'active-users' && label.textContent.includes('Users')) ||
                        (className === 'response-time' && label.textContent.includes('Response'))) {
                        element.textContent = value;
                    }
                }
            });
        }

        function updateHealthValue(className, value, healthClass) {
            const elements = document.querySelectorAll('.health-value');
            elements.forEach(element => {
                const healthItem = element.closest('.health-item');
                const label = healthItem.querySelector('.health-label');
                if (label) {
                    if ((className === 'uptime' && label.textContent.includes('Uptime')) ||
                        (className === 'queue-health' && label.textContent.includes('Queue'))) {
                        element.textContent = value;
                        if (healthClass) {
                            element.className = 'health-value ' + healthClass;
                        }
                    }
                }
            });
        }

        function flashLiveIndicator() {
            const indicator = document.querySelector('.live-indicator');
            const dot = document.querySelector('.live-dot');

            if (indicator && dot) {
                // Briefly change color to show update
                indicator.style.color = '#3b82f6';
                dot.style.background = '#3b82f6';

                setTimeout(() => {
                    indicator.style.color = '#10b981';
                    dot.style.background = '#10b981';
                }, 200);
            }

            // Add timestamp to show last update
            const statusElement = document.getElementById('connection-status');
            if (statusElement) {
                const now = new Date();
                const timeStr = now.toLocaleTimeString('en-US', {hour12: false, hour: '2-digit', minute: '2-digit', second: '2-digit'});
                statusElement.setAttribute('title', 'Last updated: ' + timeStr);
            }
        }

        // Initialize SSE connection when page loads
        window.addEventListener('load', function() {
            connectSSE();
        });

        // Cleanup on page unload
        window.addEventListener('beforeunload', function() {
            if (eventSource) {
                eventSource.close();
            }
        });

        function addJob() {
            const url = document.getElementById('url-input').value;
            if (!url) return;

            fetch('/add-job', {
                method: 'POST',
                headers: {'Content-Type': 'application/json'},
                body: JSON.stringify({url: url})
            }).then(() => {
                document.getElementById('url-input').value = '';
                // SSE will automatically update the dashboard with the new job
            });
        }

        // Navigation functions
        function showDashboard() {
            setActiveNav(0);
            document.getElementById('dashboard-view').style.display = 'block';
            document.getElementById('transcriptions-view').style.display = 'none';
            document.querySelector('.page-title').textContent = 'Transcription Dashboard';
            document.querySelector('.page-subtitle').textContent = 'Real-time video transcription monitoring';
        }

        function showTranscriptions() {
            setActiveNav(1);
            document.getElementById('dashboard-view').style.display = 'none';
            document.getElementById('transcriptions-view').style.display = 'block';
            document.querySelector('.page-title').textContent = 'All Transcriptions';
            document.querySelector('.page-subtitle').textContent = 'View and manage your transcription library';
            loadTranscriptionsView();
        }

        function showQueue() {
            setActiveNav(2);
            // Filter dashboard to show only queued jobs
            showDashboard();
            document.querySelector('.page-title').textContent = 'Processing Queue';
            document.querySelector('.page-subtitle').textContent = 'Jobs currently in the transcription queue';
            filterJobsByStatus(['pending', 'running']);
        }


        function loadTranscriptionsView() {
            const jobs = window.currentJobs || [];
            const container = document.getElementById('transcriptions-list');
            container.innerHTML = '';

            const completedJobs = jobs.filter(job => job.status === 'completed');

            if (completedJobs.length === 0) {
                container.innerHTML = '<div class="empty-state">No completed transcriptions yet. Add a job to get started!</div>';
                return;
            }

            completedJobs.forEach(job => {
                const jobElement = createTranscriptionCard(job);
                container.appendChild(jobElement);
            });
        }

        function createTranscriptionCard(job) {
            const div = document.createElement('div');
            div.className = 'transcription-card';

            const updateTime = new Date(job.update_time);
            const timeStr = updateTime.toLocaleDateString('en-US', {month: 'short', day: 'numeric', year: 'numeric'});

            div.innerHTML = ` + "`" + `
                <div class="transcription-header">
                    <div class="transcription-info">
                        <h3 class="transcription-title">` + "${job.title}" + `</h3>
                        <div class="transcription-meta">
                            <span class="meta-item">
                                <span class="meta-icon">📅</span>
                                ` + "${timeStr}" + `
                            </span>
                            <span class="meta-item">
                                <span class="meta-icon">⏱️</span>
                                ` + "${job.duration}" + `
                            </span>
                            <span class="meta-item">
                                <span class="meta-icon">📁</span>
                                ` + "${job.file_count}" + ` files
                            </span>
                            <span class="meta-item">
                                <span class="meta-icon">💾</span>
                                ` + "${job.file_size}" + `
                            </span>
                        </div>
                    </div>
                    <div class="transcription-actions">
                        <button class="action-btn view-btn" onclick="viewTranscription('` + "${job.id}" + `')">
                            <span class="btn-icon">👁️</span>
                            View
                        </button>
                        <button class="action-btn download-btn" onclick="downloadTranscription('` + "${job.id}" + `')">
                            <span class="btn-icon">⬇️</span>
                            Download
                        </button>
                    </div>
                </div>
                <div class="video-id-badge" onclick="copyToClipboard('` + "${job.video_id}" + `', event)">
                    ID: ` + "${job.video_id}" + ` <span class="copy-icon">📋</span>
                </div>
            ` + "`" + `;
            return div;
        }

        function filterJobsByStatus(statuses) {
            // This would filter the dashboard view to show only specific status jobs
            // Implementation would update the job list display
        }

        function viewTranscription(jobId) {
            window.location.href = '/transaction/' + jobId;
        }

        function downloadTranscription(jobId) {
            // Implement download functionality
            alert('Download functionality - would download transcription files for job: ' + jobId);
        }

        function filterTranscriptions(period) {
            // Update active filter button
            document.querySelectorAll('.filter-btn').forEach(btn => btn.classList.remove('active'));
            event.target.classList.add('active');

            const jobs = window.currentJobs || [];
            const now = new Date();
            let filteredJobs = jobs.filter(job => job.status === 'completed');

            if (period !== 'all') {
                filteredJobs = filteredJobs.filter(job => {
                    const jobDate = new Date(job.update_time);
                    switch (period) {
                        case 'today':
                            return jobDate.toDateString() === now.toDateString();
                        case 'week':
                            const weekAgo = new Date(now.getTime() - 7 * 24 * 60 * 60 * 1000);
                            return jobDate >= weekAgo;
                        case 'month':
                            const monthAgo = new Date(now.getTime() - 30 * 24 * 60 * 60 * 1000);
                            return jobDate >= monthAgo;
                        default:
                            return true;
                    }
                });
            }

            const container = document.getElementById('transcriptions-list');
            container.innerHTML = '';

            if (filteredJobs.length === 0) {
                container.innerHTML = '<div class="empty-state">No transcriptions found for this time period.</div>';
                return;
            }

            filteredJobs.forEach(job => {
                const jobElement = createTranscriptionCard(job);
                container.appendChild(jobElement);
            });
        }

        function setActiveNav(index) {
            const navItems = document.querySelectorAll('.nav-item');
            navItems.forEach((item, i) => {
                if (i === index) {
                    item.classList.add('active');
                } else {
                    item.classList.remove('active');
                }
            });
        }

        // Copy to clipboard function
        function copyToClipboard(text, event) {
            event.stopPropagation();
            navigator.clipboard.writeText(text).then(function() {
                // Visual feedback
                const element = event.target.closest('.video-id');
                const originalBg = element.style.backgroundColor;
                element.style.backgroundColor = '#10b981';
                element.style.color = 'white';
                setTimeout(() => {
                    element.style.backgroundColor = originalBg;
                    element.style.color = '#6b7280';
                }, 200);
            }).catch(function(err) {
                console.error('Could not copy text: ', err);
                // Fallback for older browsers
                const textArea = document.createElement('textarea');
                textArea.value = text;
                document.body.appendChild(textArea);
                textArea.select();
                document.execCommand('copy');
                document.body.removeChild(textArea);
            });
        }

        // Job detail function
        function showJobDetails(jobId) {
            window.location.href = '/transaction/' + jobId;
        }

        // Live reload functionality
        (function() {
            let lastModified = Date.now();

            function checkForUpdates() {
                fetch('/api/reload-check')
                    .then(response => response.json())
                    .then(data => {
                        if (data.modified > lastModified) {
                            console.log('🔄 File changes detected - refreshing page...');
                            window.location.reload();
                        }
                        lastModified = data.modified;
                    })
                    .catch(() => {
                        // Silently fail - server might be restarting
                    });
            }

            // Check for updates every 500ms
            setInterval(checkForUpdates, 500);
            console.log('🔄 Live reload monitoring active');
        })();
    </script>
</body>
</html>
`

const transactionDetailHTML = `
<!DOCTYPE html>
<html>
<head>
    <title>Transaction Details - OmniTranscripts</title>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <link href="https://fonts.googleapis.com/css2?family=Inter:wght@300;400;500;600;700&display=swap" rel="stylesheet">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }

        body {
            font-family: 'Inter', -apple-system, BlinkMacSystemFont, sans-serif;
            background: #f5f7fa;
            min-height: 100vh;
        }

        .header {
            background: #ffffff;
            border-bottom: 1px solid #e5e7eb;
            padding: 16px 32px;
            display: flex;
            align-items: center;
            justify-content: space-between;
        }

        .header-left {
            display: flex;
            align-items: center;
            gap: 16px;
        }

        .back-button {
            background: #f3f4f6;
            border: none;
            padding: 8px 16px;
            border-radius: 8px;
            cursor: pointer;
            font-size: 14px;
            text-decoration: none;
            color: #374151;
            display: flex;
            align-items: center;
            gap: 8px;
        }

        .back-button:hover {
            background: #e5e7eb;
        }

        .header-title {
            font-size: 24px;
            font-weight: 600;
            color: #111827;
        }

        .container {
            max-width: 1200px;
            margin: 0 auto;
            padding: 32px;
        }

        .detail-card {
            background: #ffffff;
            border-radius: 16px;
            border: 1px solid #e5e7eb;
            padding: 32px;
            margin-bottom: 24px;
        }

        .status-badge {
            display: inline-block;
            padding: 6px 12px;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 500;
            text-transform: uppercase;
            letter-spacing: 0.05em;
        }

        .status-completed { background: #d1fae5; color: #065f46; }
        .status-failed { background: #fee2e2; color: #991b1b; }
        .status-queued { background: #fef3c7; color: #92400e; }
        .status-running { background: #dbeafe; color: #1e40af; }

        .detail-grid {
            display: grid;
            grid-template-columns: repeat(auto-fit, minmax(300px, 1fr));
            gap: 24px;
            margin-top: 24px;
        }

        .detail-section {
            background: #f9fafb;
            border-radius: 12px;
            padding: 20px;
        }

        .detail-section h3 {
            font-size: 16px;
            font-weight: 600;
            color: #111827;
            margin-bottom: 16px;
        }

        .detail-row {
            display: flex;
            justify-content: space-between;
            align-items: center;
            padding: 8px 0;
            border-bottom: 1px solid #e5e7eb;
        }

        .detail-row:last-child {
            border-bottom: none;
        }

        .detail-label {
            font-size: 14px;
            color: #6b7280;
            font-weight: 500;
        }

        .detail-value {
            font-size: 14px;
            color: #111827;
            font-weight: 500;
            text-align: right;
        }

        .video-id {
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            font-size: 12px;
            color: #6b7280;
            background: #f3f4f6;
            padding: 4px 8px;
            border-radius: 6px;
            cursor: pointer;
            display: inline-flex;
            align-items: center;
            gap: 4px;
        }

        .video-id:hover {
            background: #e5e7eb;
        }

        .progress-bar {
            width: 100%;
            height: 8px;
            background: #e5e7eb;
            border-radius: 4px;
            overflow: hidden;
        }

        .progress-fill {
            height: 100%;
            background: linear-gradient(90deg, #10b981, #059669);
            transition: width 0.3s ease;
        }

        .url-link {
            color: #3b82f6;
            text-decoration: none;
            word-break: break-all;
        }

        .url-link:hover {
            text-decoration: underline;
        }

        @media (max-width: 768px) {
            .container {
                padding: 16px;
            }

            .detail-card {
                padding: 20px;
            }

            .detail-grid {
                grid-template-columns: 1fr;
            }
        }
    </style>
</head>
<body>
    <div class="header">
        <div class="header-left">
            <a href="/" class="back-button">
                ← Back to Dashboard
            </a>
            <h1 class="header-title">Transaction Details</h1>
        </div>
    </div>

    <div class="container">
        <div class="detail-card">
            <div style="display: flex; justify-content: space-between; align-items: flex-start; margin-bottom: 24px;">
                <div>
                    <h2 style="font-size: 20px; font-weight: 600; color: #111827; margin-bottom: 8px;">{{.Job.Title}}</h2>
                    <span class="video-id" onclick="copyToClipboard('{{.Job.VideoID}}')" title="Click to copy Video ID">
                        {{.Job.VideoID}} <span>📋</span>
                    </span>
                </div>
                <span class="status-badge status-{{.Job.Status}}">{{.Job.Status}}</span>
            </div>

            <div class="detail-grid">
                <div class="detail-section">
                    <h3>Job Information</h3>
                    <div class="detail-row">
                        <span class="detail-label">Job ID</span>
                        <span class="detail-value">{{.Job.ID}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Video ID</span>
                        <span class="detail-value">{{.Job.VideoID}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Status</span>
                        <span class="detail-value">{{.Job.StatusText}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Progress</span>
                        <div class="detail-value" style="width: 100px;">
                            <div class="progress-bar">
                                <div class="progress-fill" style="width: {{.Job.Progress}}%"></div>
                            </div>
                            <div style="font-size: 12px; text-align: center; margin-top: 4px;">{{.Job.Progress}}%</div>
                        </div>
                    </div>
                </div>

                <div class="detail-section">
                    <h3>Timing</h3>
                    <div class="detail-row">
                        <span class="detail-label">Started</span>
                        <span class="detail-value">{{.Job.StartTime.Format "Jan 2, 2006 15:04"}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Last Updated</span>
                        <span class="detail-value">{{.Job.UpdateTime.Format "Jan 2, 2006 15:04"}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Duration</span>
                        <span class="detail-value">{{.Job.Duration}}</span>
                    </div>
                </div>

                <div class="detail-section">
                    <h3>Output</h3>
                    <div class="detail-row">
                        <span class="detail-label">File Count</span>
                        <span class="detail-value">{{.Job.FileCount}} files</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Total Size</span>
                        <span class="detail-value">{{.Job.FileSize}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Output Directory</span>
                        <span class="detail-value">{{.Job.OutputDir}}</span>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Log File</span>
                        <span class="detail-value">{{.Job.LogFile}}</span>
                    </div>
                </div>

                <div class="detail-section">
                    <h3>Source</h3>
                    <div class="detail-row">
                        <span class="detail-label">URL</span>
                        <div class="detail-value">
                            <a href="{{.Job.URL}}" target="_blank" class="url-link">{{.Job.URL}}</a>
                        </div>
                    </div>
                    <div class="detail-row">
                        <span class="detail-label">Category</span>
                        <span class="detail-value">{{.Job.CategoryIcon}} {{.Job.CategoryClass}}</span>
                    </div>
                </div>
            </div>

            {{if .Job.Transcript}}
            <!-- Transcript Section -->
            <div class="detail-card">
                <div style="display: flex; justify-content: space-between; align-items: center; margin-bottom: 24px;">
                    <h3 style="font-size: 18px; font-weight: 600; color: #111827; margin: 0;">📝 Transcript</h3>
                    <div style="display: flex; gap: 12px;">
                        <button onclick="copyTranscript()" class="action-btn copy-btn" title="Copy transcript">
                            📋 Copy
                        </button>
                        <button onclick="downloadTranscript('txt')" class="action-btn download-btn" title="Download as text">
                            📄 TXT
                        </button>
                        <button onclick="downloadTranscript('srt')" class="action-btn download-btn" title="Download as SRT">
                            🎬 SRT
                        </button>
                        <button onclick="downloadTranscript('vtt')" class="action-btn download-btn" title="Download as VTT">
                            📺 VTT
                        </button>
                        <button onclick="downloadTranscript('json')" class="action-btn download-btn" title="Download as JSON">
                            🔧 JSON
                        </button>
                    </div>
                </div>

                <!-- Search Box -->
                <div style="margin-bottom: 20px;">
                    <input type="text" id="transcript-search" placeholder="🔍 Search in transcript..."
                           style="width: 100%; padding: 12px; border: 1px solid #e5e7eb; border-radius: 8px; font-size: 14px;"
                           onkeyup="searchTranscript()" />
                </div>

                <!-- Transcript Content -->
                <div class="transcript-container">
                    {{if .Job.Segments}}
                    <!-- Timestamped Segments -->
                    <div class="transcript-segments" id="transcript-segments">
                        {{range .Job.Segments}}
                        <div class="transcript-segment" data-start="{{.Start}}" data-end="{{.End}}">
                            <span class="timestamp" onclick="seekToTime({{.Start}})">
                                {{formatTime .Start}}
                            </span>
                            <span class="segment-text">{{.Text}}</span>
                        </div>
                        {{end}}
                    </div>
                    {{else}}
                    <!-- Plain Text Transcript -->
                    <div class="transcript-text" id="transcript-text">
                        {{.Job.Transcript}}
                    </div>
                    {{end}}
                </div>
            </div>
            {{else}}
            <!-- No Transcript Available -->
            {{if eq .Job.Status "completed"}}
            <div class="detail-card">
                <div style="text-align: center; padding: 40px; color: #6b7280;">
                    <div style="font-size: 48px; margin-bottom: 16px;">📄</div>
                    <h3 style="margin: 0 0 8px 0; color: #374151;">No Transcript Available</h3>
                    <p style="margin: 0;">The transcription completed but no transcript content was found.</p>
                </div>
            </div>
            {{else if eq .Job.Status "failed"}}
            <div class="detail-card">
                <div style="text-align: center; padding: 40px; color: #ef4444;">
                    <div style="font-size: 48px; margin-bottom: 16px;">❌</div>
                    <h3 style="margin: 0 0 8px 0; color: #dc2626;">Transcription Failed</h3>
                    <p style="margin: 0; color: #6b7280;">The transcription process encountered an error.</p>
                    {{if .Job.LogFile}}
                    <button onclick="viewLogs()" style="margin-top: 16px; padding: 8px 16px; background: #f3f4f6; border: 1px solid #d1d5db; border-radius: 6px; cursor: pointer;">
                        View Error Log
                    </button>
                    {{end}}
                </div>
            </div>
            {{else}}
            <div class="detail-card">
                <div style="text-align: center; padding: 40px; color: #6b7280;">
                    <div style="font-size: 48px; margin-bottom: 16px;">⏳</div>
                    <h3 style="margin: 0 0 8px 0; color: #374151;">Transcription In Progress</h3>
                    <p style="margin: 0;">The transcript will appear here once processing is complete.</p>
                </div>
            </div>
            {{end}}
            {{end}}
        </div>
    </div>

    <style>
        .action-btn {
            padding: 8px 12px;
            border: 1px solid #d1d5db;
            border-radius: 6px;
            background: white;
            color: #374151;
            cursor: pointer;
            font-size: 12px;
            font-weight: 500;
            transition: all 0.2s;
        }
        .action-btn:hover {
            background: #f9fafb;
            border-color: #9ca3af;
        }
        .copy-btn:hover {
            background: #dbeafe;
            border-color: #3b82f6;
            color: #1d4ed8;
        }
        .download-btn:hover {
            background: #d1fae5;
            border-color: #10b981;
            color: #047857;
        }
        .transcript-container {
            max-height: 600px;
            overflow-y: auto;
            border: 1px solid #e5e7eb;
            border-radius: 8px;
            padding: 16px;
            background: #fafafa;
        }
        .transcript-segments .transcript-segment {
            margin-bottom: 12px;
            padding: 8px;
            border-radius: 6px;
            transition: background-color 0.2s;
        }
        .transcript-segment:hover {
            background: #f3f4f6;
        }
        .transcript-segment.highlight {
            background: #fef3c7;
            border: 1px solid #f59e0b;
        }
        .timestamp {
            display: inline-block;
            width: 80px;
            color: #6b7280;
            font-size: 12px;
            font-family: 'Monaco', 'Menlo', 'Consolas', monospace;
            cursor: pointer;
            margin-right: 12px;
            text-align: right;
        }
        .timestamp:hover {
            color: #3b82f6;
            text-decoration: underline;
        }
        .segment-text {
            color: #111827;
            line-height: 1.5;
        }
        .transcript-text {
            color: #111827;
            line-height: 1.6;
            white-space: pre-wrap;
        }
    </style>

    <script>
        function copyToClipboard(text) {
            navigator.clipboard.writeText(text).then(function() {
                // Visual feedback
                const elements = document.querySelectorAll('.video-id');
                elements.forEach(element => {
                    const originalBg = element.style.backgroundColor;
                    element.style.backgroundColor = '#10b981';
                    element.style.color = 'white';
                    setTimeout(() => {
                        element.style.backgroundColor = originalBg;
                        element.style.color = '#6b7280';
                    }, 200);
                });
            });
        }

        // Transcript functionality
        function copyTranscript() {
            const transcriptElement = document.getElementById('transcript-text') || document.getElementById('transcript-segments');
            let text = '';

            if (document.getElementById('transcript-segments')) {
                // Copy timestamped segments
                const segments = document.querySelectorAll('.transcript-segment');
                segments.forEach(segment => {
                    const timestamp = segment.querySelector('.timestamp').textContent;
                    const segmentText = segment.querySelector('.segment-text').textContent;
                    text += '[' + timestamp + '] ' + segmentText + '\n';
                });
            } else if (document.getElementById('transcript-text')) {
                // Copy plain text
                text = document.getElementById('transcript-text').textContent;
            }

            navigator.clipboard.writeText(text).then(() => {
                // Visual feedback
                const btn = event.target;
                const originalText = btn.textContent;
                btn.textContent = '✅ Copied!';
                btn.style.background = '#d1fae5';
                btn.style.borderColor = '#10b981';
                setTimeout(() => {
                    btn.textContent = originalText;
                    btn.style.background = '';
                    btn.style.borderColor = '';
                }, 2000);
            });
        }

        function downloadTranscript(format) {
            const jobId = window.location.pathname.split('/').pop();
            window.open('/download/' + jobId + '/' + format, '_blank');
        }

        function searchTranscript() {
            const searchTerm = document.getElementById('transcript-search').value.toLowerCase();
            const segments = document.querySelectorAll('.transcript-segment');
            const transcriptText = document.getElementById('transcript-text');

            if (segments.length > 0) {
                // Search in timestamped segments
                segments.forEach(segment => {
                    const text = segment.querySelector('.segment-text').textContent.toLowerCase();
                    if (searchTerm === '' || text.includes(searchTerm)) {
                        segment.style.display = 'block';
                        if (searchTerm !== '') {
                            segment.classList.add('highlight');
                        } else {
                            segment.classList.remove('highlight');
                        }
                    } else {
                        segment.style.display = 'none';
                        segment.classList.remove('highlight');
                    }
                });
            } else if (transcriptText) {
                // Simple highlight for plain text (basic implementation)
                const originalText = transcriptText.getAttribute('data-original') || transcriptText.textContent;
                if (!transcriptText.getAttribute('data-original')) {
                    transcriptText.setAttribute('data-original', originalText);
                }

                if (searchTerm === '') {
                    transcriptText.innerHTML = originalText;
                } else {
                    const regex = new RegExp('(' + searchTerm + ')', 'gi');
                    transcriptText.innerHTML = originalText.replace(regex, '<mark style="background: #fef3c7; padding: 2px;">$1</mark>');
                }
            }
        }

        function seekToTime(seconds) {
            // This would integrate with a video player if available
            // For now, just show a tooltip with the timestamp
            alert('Seeking to ' + formatTime(seconds));
        }

        function formatTime(seconds) {
            const hours = Math.floor(seconds / 3600);
            const minutes = Math.floor((seconds % 3600) / 60);
            const secs = Math.floor(seconds % 60);

            if (hours > 0) {
                return hours + ':' + minutes.toString().padStart(2, '0') + ':' + secs.toString().padStart(2, '0');
            } else {
                return minutes + ':' + secs.toString().padStart(2, '0');
            }
        }

        function viewLogs() {
            const jobId = window.location.pathname.split('/').pop();
            window.open('/logs/' + jobId, '_blank');
        }
    </script>
</body>
</html>
`

type DashboardData struct {
	Jobs         []Job
	TotalJobs    int
	RunningJobs  int
	CompletedJobs int
	FailedJobs   int
	QueuedJobs   int

	// Percentages for progress bars
	CompletedPercentage int
	RunningPercentage   int
	FailedPercentage    int
	QueuedPercentage    int

	// Performance metrics
	SuccessRate         float64
	AvgProcessingTime   string
	TotalDuration       string
	StorageUsed         string

	// API Usage metrics
	JobsToday      int
	JobsThisWeek   int
	ActiveUsers    int
	AvgResponseTime int

	// System health
	Uptime           string
	QueueHealth      string
	QueueHealthClass string
	// Business insights for monetization
	RevenueToday       float64
	RevenueGrowth      float64
	AvgRevenuePerJob   float64
	JobsYesterday      int
}

func main() {
	// Initialize file modification time for live reload
	if stat, err := os.Stat("web-dashboard.go"); err == nil {
		fileMutex.Lock()
		fileModTime = stat.ModTime()
		fileMutex.Unlock()
	}

	// Start file watcher for live reload
	go func() {
		for {
			time.Sleep(100 * time.Millisecond)
			if stat, err := os.Stat("web-dashboard.go"); err == nil {
				fileMutex.RLock()
				lastMod := fileModTime
				fileMutex.RUnlock()

				if stat.ModTime().After(lastMod) {
					fileMutex.Lock()
					fileModTime = stat.ModTime()
					fileMutex.Unlock()
				}
			}
		}
	}()

	http.HandleFunc("/", dashboardHandler)
	http.HandleFunc("/api/jobs", jobsHandler)
	http.HandleFunc("/add-job", addJobHandler)
	http.HandleFunc("/transaction/", transactionDetailHandler)
	http.HandleFunc("/download/", downloadHandler)
	http.HandleFunc("/logs/", logsHandler)
	http.HandleFunc("/events", sseHandler)
	http.HandleFunc("/api/reload-check", reloadCheckHandler)
	http.HandleFunc("/demo/add-transcript/", demoAddTranscriptHandler)

	port := findAvailablePort(8765)

	fmt.Println("🚀 OmniTranscripts Command Center starting...")
	fmt.Printf("🌐 Dashboard: http://localhost:%d\n", port)
	fmt.Printf("📊 API: http://localhost:%d/api/jobs\n", port)

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func findAvailablePort(startPort int) int {
	for port := startPort; port < startPort+100; port++ {
		ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
		if err == nil {
			ln.Close()
			return port
		}
	}
	log.Fatal("No available ports found")
	return 0
}

func dashboardHandler(w http.ResponseWriter, r *http.Request) {
	jobs := loadJobs()
	updateJobStatuses(jobs)

	data := DashboardData{
		Jobs:      jobs,
		TotalJobs: len(jobs),
	}

	// Calculate basic job counts
	for _, job := range jobs {
		switch job.Status {
		case "downloading", "extracting", "transcribing":
			data.RunningJobs++
		case "completed":
			data.CompletedJobs++
		case "failed":
			data.FailedJobs++
		case "queued":
			data.QueuedJobs++
		}
	}

	// Calculate percentages for progress bars
	if data.TotalJobs > 0 {
		data.CompletedPercentage = (data.CompletedJobs * 100) / data.TotalJobs
		data.RunningPercentage = (data.RunningJobs * 100) / data.TotalJobs
		data.FailedPercentage = (data.FailedJobs * 100) / data.TotalJobs
		data.QueuedPercentage = (data.QueuedJobs * 100) / data.TotalJobs
	}

	// Calculate performance metrics
	calculatePerformanceMetrics(&data, jobs)

	// Calculate API usage metrics
	calculateAPIUsageMetrics(&data, jobs)

	// Calculate system health metrics
	calculateSystemHealthMetrics(&data)

	// Calculate business insights
	calculateBusinessMetrics(&data, jobs)

	tmpl := template.Must(template.New("dashboard").Parse(dashboardHTML))
	tmpl.Execute(w, data)
}

func calculatePerformanceMetrics(data *DashboardData, jobs []Job) {
	// Success rate
	if data.TotalJobs > 0 {
		data.SuccessRate = float64(data.CompletedJobs) / float64(data.TotalJobs) * 100
	}

	// Average processing time and total duration
	var totalProcessingSeconds int64
	var totalVideoSeconds int64
	var completedJobs int

	for _, job := range jobs {
		if job.Status == "completed" && !job.StartTime.IsZero() && !job.UpdateTime.IsZero() {
			processingTime := job.UpdateTime.Sub(job.StartTime)
			totalProcessingSeconds += int64(processingTime.Seconds())
			completedJobs++
		}

		// Parse duration to calculate total video time processed
		if duration := parseDurationToSeconds(job.Duration); duration > 0 {
			totalVideoSeconds += duration
		}
	}

	if completedJobs > 0 {
		avgSeconds := totalProcessingSeconds / int64(completedJobs)
		data.AvgProcessingTime = formatDuration(time.Duration(avgSeconds) * time.Second)
	} else {
		data.AvgProcessingTime = "N/A"
	}

	// Total duration of video content processed
	data.TotalDuration = formatDuration(time.Duration(totalVideoSeconds) * time.Second)

	// Calculate total storage used
	var totalStorage int64
	for _, job := range jobs {
		if job.Status == "completed" {
			outputDir := fmt.Sprintf("transcripts/%s", job.VideoID)
			if files, err := os.ReadDir(outputDir); err == nil {
				for _, file := range files {
					if fileStat, err := os.Stat(fmt.Sprintf("%s/%s", outputDir, file.Name())); err == nil {
						totalStorage += fileStat.Size()
					}
				}
			}
		}
	}
	data.StorageUsed = formatFileSize(totalStorage)
}

func calculateAPIUsageMetrics(data *DashboardData, jobs []Job) {
	now := time.Now()
	today := now.Truncate(24 * time.Hour)
	weekAgo := now.AddDate(0, 0, -7)

	for _, job := range jobs {
		// Jobs today
		if job.StartTime.After(today) {
			data.JobsToday++
		}

		// Jobs this week
		if job.StartTime.After(weekAgo) {
			data.JobsThisWeek++
		}
	}

	// Simulate active users (could be enhanced with actual user tracking)
	data.ActiveUsers = data.JobsToday / 3 // Rough estimate
	if data.ActiveUsers < 1 && data.JobsToday > 0 {
		data.ActiveUsers = 1
	}

	// Simulate average response time (could be enhanced with actual timing)
	data.AvgResponseTime = 250 + (data.RunningJobs * 50) // Increases with load
}

func calculateSystemHealthMetrics(data *DashboardData) {
	// Calculate uptime
	uptime := time.Since(startTime)
	data.Uptime = formatDuration(uptime)

	// Queue health assessment
	if data.QueuedJobs == 0 && data.RunningJobs <= 2 {
		data.QueueHealth = "Healthy"
		data.QueueHealthClass = "healthy"
	} else if data.QueuedJobs <= 3 && data.RunningJobs <= 5 {
		data.QueueHealth = "Moderate"
		data.QueueHealthClass = "warning"
	} else {
		data.QueueHealth = "High Load"
		data.QueueHealthClass = "critical"
	}
}

func parseDurationToSeconds(duration string) int64 {
	// Parse duration strings like "04:34:06" or "02:35"
	parts := strings.Split(duration, ":")
	if len(parts) < 2 {
		return 0
	}

	var seconds int64
	for i, part := range parts {
		if val, err := strconv.ParseInt(part, 10, 64); err == nil {
			switch len(parts) - i - 1 {
			case 0: // seconds
				seconds += val
			case 1: // minutes
				seconds += val * 60
			case 2: // hours
				seconds += val * 3600
			}
		}
	}
	return seconds
}

func jobsHandler(w http.ResponseWriter, r *http.Request) {
	jobs := loadJobs()
	updateJobStatuses(jobs)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(jobs)
}

func addJobHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", 405)
		return
	}

	var req struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	videoID := extractVideoID(req.URL)
	jobID := generateJobID()

	job := Job{
		ID:         jobID,
		VideoID:    videoID,
		URL:        req.URL,
		Title:      "Loading...",
		Status:     "queued",
		Progress:   0,
		StartTime:  time.Now(),
		UpdateTime: time.Now(),
		LogFile:    fmt.Sprintf("logs/%s.log", jobID),
		OutputDir:  fmt.Sprintf("transcripts/%s", videoID),
		Duration:   "00:00",
		FileCount:  0,
		FileSize:   "0 KB",
		CategoryClass: "entertainment",
		CategoryIcon:  "🎬",
		StatusText:    "Queued for processing",
	}

	jobs := loadJobs()
	jobs = append(jobs, job)
	saveJobs(jobs)

	w.WriteHeader(200)
}

func transactionDetailHandler(w http.ResponseWriter, r *http.Request) {
	// Extract job ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/transaction/")
	jobID := strings.TrimSuffix(path, "/")

	if jobID == "" {
		http.Error(w, "Job ID required", 400)
		return
	}

	jobs := loadJobs()
	var job *Job
	for i := range jobs {
		if jobs[i].ID == jobID {
			job = &jobs[i]
			break
		}
	}

	if job == nil {
		http.Error(w, "Job not found", 404)
		return
	}

	updateJobStatuses(jobs)

	// Merge demo transcript data if available
	if demoData, exists := demoTranscripts[jobID]; exists {
		job.Transcript = demoData.Transcript
		job.Segments = demoData.Segments
	}

	// Template functions
	funcMap := template.FuncMap{
		"formatTime": func(seconds float64) string {
			hours := int(seconds) / 3600
			minutes := (int(seconds) % 3600) / 60
			secs := int(seconds) % 60

			if hours > 0 {
				return fmt.Sprintf("%d:%02d:%02d", hours, minutes, secs)
			}
			return fmt.Sprintf("%d:%02d", minutes, secs)
		},
	}

	tmpl, err := template.New("transaction").Funcs(funcMap).Parse(transactionDetailHTML)
	if err != nil {
		http.Error(w, "Template error", 500)
		return
	}

	data := struct {
		Job *Job
	}{
		Job: job,
	}

	w.Header().Set("Content-Type", "text/html")
	tmpl.Execute(w, data)
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /download/{jobId}/{format}
	pathParts := strings.Split(strings.TrimPrefix(r.URL.Path, "/download/"), "/")
	if len(pathParts) != 2 {
		http.Error(w, "Invalid download URL format. Use /download/{jobId}/{format}", 400)
		return
	}

	jobID := pathParts[0]
	format := pathParts[1]

	// Validate format
	validFormats := map[string]string{
		"txt":  "text/plain",
		"srt":  "application/x-subrip",
		"vtt":  "text/vtt",
		"json": "application/json",
	}

	mimeType, validFormat := validFormats[format]
	if !validFormat {
		http.Error(w, "Invalid format. Supported: txt, srt, vtt, json", 400)
		return
	}

	// Find job
	jobs := loadJobs()
	var job *Job
	for i := range jobs {
		if jobs[i].ID == jobID {
			job = &jobs[i]
			break
		}
	}

	if job == nil {
		http.Error(w, "Job not found", 404)
		return
	}

	// Merge demo transcript data if available
	if demoData, exists := demoTranscripts[jobID]; exists {
		job.Transcript = demoData.Transcript
		job.Segments = demoData.Segments
	}

	if job.Transcript == "" {
		http.Error(w, "No transcript available for this job", 404)
		return
	}

	// Generate content based on format
	var content string
	var filename string

	switch format {
	case "txt":
		content = job.Transcript
		filename = fmt.Sprintf("transcript_%s.txt", job.VideoID)

	case "srt":
		content = generateSRT(job.Segments, job.Transcript)
		filename = fmt.Sprintf("transcript_%s.srt", job.VideoID)

	case "vtt":
		content = generateVTT(job.Segments, job.Transcript)
		filename = fmt.Sprintf("transcript_%s.vtt", job.VideoID)

	case "json":
		jsonData := map[string]interface{}{
			"job_id":     job.ID,
			"video_id":   job.VideoID,
			"title":      job.Title,
			"url":        job.URL,
			"transcript": job.Transcript,
			"segments":   job.Segments,
			"duration":   job.Duration,
			"created_at": job.StartTime,
		}
		contentBytes, _ := json.MarshalIndent(jsonData, "", "  ")
		content = string(contentBytes)
		filename = fmt.Sprintf("transcript_%s.json", job.VideoID)
	}

	// Set headers
	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"", filename))
	w.Header().Set("Content-Length", fmt.Sprintf("%d", len(content)))

	// Write content
	w.Write([]byte(content))
}

func generateSRT(segments []TranscriptSegment, fallbackText string) string {
	if len(segments) == 0 {
		// Fallback for plain text
		return fmt.Sprintf("1\n00:00:00,000 --> 99:59:59,999\n%s\n", fallbackText)
	}

	var srt strings.Builder
	for i, segment := range segments {
		start := formatSRTTime(segment.Start)
		end := formatSRTTime(segment.End)
		srt.WriteString(fmt.Sprintf("%d\n%s --> %s\n%s\n\n", i+1, start, end, segment.Text))
	}
	return srt.String()
}

func generateVTT(segments []TranscriptSegment, fallbackText string) string {
	var vtt strings.Builder
	vtt.WriteString("WEBVTT\n\n")

	if len(segments) == 0 {
		// Fallback for plain text
		vtt.WriteString("00:00:00.000 --> 99:59:59.999\n")
		vtt.WriteString(fallbackText)
		vtt.WriteString("\n")
		return vtt.String()
	}

	for _, segment := range segments {
		start := formatVTTTime(segment.Start)
		end := formatVTTTime(segment.End)
		vtt.WriteString(fmt.Sprintf("%s --> %s\n%s\n\n", start, end, segment.Text))
	}
	return vtt.String()
}

func formatSRTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d,%03d", hours, minutes, secs, millis)
}

func formatVTTTime(seconds float64) string {
	hours := int(seconds) / 3600
	minutes := (int(seconds) % 3600) / 60
	secs := int(seconds) % 60
	millis := int((seconds - float64(int(seconds))) * 1000)
	return fmt.Sprintf("%02d:%02d:%02d.%03d", hours, minutes, secs, millis)
}

func logsHandler(w http.ResponseWriter, r *http.Request) {
	// Parse URL: /logs/{jobId}
	jobID := strings.TrimPrefix(r.URL.Path, "/logs/")
	if jobID == "" {
		http.Error(w, "Job ID required", 400)
		return
	}

	// Find job
	jobs := loadJobs()
	var job *Job
	for i := range jobs {
		if jobs[i].ID == jobID {
			job = &jobs[i]
			break
		}
	}

	if job == nil {
		http.Error(w, "Job not found", 404)
		return
	}

	if job.LogFile == "" {
		http.Error(w, "No log file available for this job", 404)
		return
	}

	// Try to read the log file
	content, err := os.ReadFile(job.LogFile)
	if err != nil {
		http.Error(w, "Log file not found or could not be read", 404)
		return
	}

	// Set headers
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf("inline; filename=\"%s_log.txt\"", job.VideoID))

	// Write content
	w.Write(content)
}

func sseHandler(w http.ResponseWriter, r *http.Request) {
	// Set comprehensive CORS headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Cache-Control, Content-Type, Authorization")
	w.Header().Set("Access-Control-Expose-Headers", "Content-Type")

	// Notify client connection
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming unsupported", http.StatusInternalServerError)
		return
	}

	// Send initial data
	jobs := loadJobs()
	updateJobStatuses(jobs)

	// Calculate dashboard data
	data := calculateDashboardData(jobs)

	// Send jobs data
	jobsJSON, _ := json.Marshal(jobs)
	fmt.Fprintf(w, "event: jobs\ndata: %s\n\n", string(jobsJSON))

	// Send dashboard stats
	statsJSON, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: stats\ndata: %s\n\n", string(statsJSON))

	flusher.Flush()

	// Keep connection alive and send updates every 3 seconds
	ticker := time.NewTicker(3 * time.Second)
	defer ticker.Stop()

	// Channel to detect client disconnect
	clientGone := r.Context().Done()

	for {
		select {
		case <-clientGone:
			return
		case <-ticker.C:
			// Load fresh data
			jobs := loadJobs()
			updateJobStatuses(jobs)
			data := calculateDashboardData(jobs)

			// Send updated jobs
			jobsJSON, _ := json.Marshal(jobs)
			fmt.Fprintf(w, "event: jobs\ndata: %s\n\n", string(jobsJSON))

			// Send updated stats
			statsJSON, _ := json.Marshal(data)
			fmt.Fprintf(w, "event: stats\ndata: %s\n\n", string(statsJSON))

			flusher.Flush()
		}
	}
}

func calculateDashboardData(jobs []Job) DashboardData {
	// Initialize dashboard data
	data := DashboardData{
		Jobs: jobs,
		TotalJobs: len(jobs),
	}

	// Calculate job statistics
	for _, job := range jobs {
		switch job.Status {
		case "completed":
			data.CompletedJobs++
		case "failed":
			data.FailedJobs++
		case "queued":
			data.QueuedJobs++
		case "downloading", "extracting", "transcribing", "running":
			data.RunningJobs++
		}
	}

	// Calculate percentages
	if data.TotalJobs > 0 {
		data.CompletedPercentage = int(float64(data.CompletedJobs) / float64(data.TotalJobs) * 100)
		data.FailedPercentage = int(float64(data.FailedJobs) / float64(data.TotalJobs) * 100)
		data.QueuedPercentage = int(float64(data.QueuedJobs) / float64(data.TotalJobs) * 100)
		data.RunningPercentage = int(float64(data.RunningJobs) / float64(data.TotalJobs) * 100)
	}

	// Calculate performance metrics
	calculatePerformanceMetrics(&data, jobs)

	// Calculate API usage metrics
	calculateAPIUsageMetrics(&data, jobs)

	// Calculate system health metrics
	calculateSystemHealthMetrics(&data)

	// Calculate business metrics
	calculateBusinessMetrics(&data, jobs)

	return data
}

// Helper functions (same as before)
func loadJobs() []Job {
	data, err := os.ReadFile("jobs.json")
	if err != nil {
		return []Job{}
	}

	var jobs []Job
	json.Unmarshal(data, &jobs)

	// Set default values for new fields
	for i := range jobs {
		if jobs[i].CategoryClass == "" {
			jobs[i].CategoryClass = "entertainment"
		}
		if jobs[i].CategoryIcon == "" {
			jobs[i].CategoryIcon = "🎬"
		}
		if jobs[i].StatusText == "" {
			updateStatusText(&jobs[i])
		}
	}

	return jobs
}

func saveJobs(jobs []Job) {
	os.MkdirAll("logs", 0755)
	data, _ := json.MarshalIndent(jobs, "", "  ")
	os.WriteFile("jobs.json", data, 0644)
}

func updateJobStatuses(jobs []Job) {
	for i := range jobs {
		updateJobStatus(&jobs[i])
	}
	saveJobs(jobs)
}

func updateJobStatus(job *Job) {
	// Load metadata if available
	metadataPath := fmt.Sprintf("transcripts/%s/metadata.json", job.VideoID)
	if metadataData, err := os.ReadFile(metadataPath); err == nil {
		var metadata map[string]interface{}
		if json.Unmarshal(metadataData, &metadata) == nil {
			if title, ok := metadata["title"].(string); ok && job.Title == "Loading..." {
				job.Title = title
			}
			if url, ok := metadata["url"].(string); ok && job.URL == "" {
				job.URL = url
			}
		}
	}

	// If title is still "Loading...", try to fetch it directly with yt-dlp
	if job.Title == "Loading..." && job.URL != "" {
		if title := fetchVideoTitle(job.URL); title != "" {
			job.Title = title
		}
	}

	// Check if transcription process is running
	cmd := exec.Command("pgrep", "-f", fmt.Sprintf("transcribe.*%s", job.VideoID))
	if output, err := cmd.Output(); err == nil && len(strings.TrimSpace(string(output))) > 0 {
		status, progress := parseJobProgress(job)
		job.Status = status
		job.Progress = progress
		job.UpdateTime = time.Now()
		updateJobStats(job)
		updateStatusText(job)
		return
	}

	// Check if completed
	outputDir := fmt.Sprintf("transcripts/%s", job.VideoID)
	if files, err := os.ReadDir(outputDir); err == nil && len(files) > 0 {
		if job.Status != "completed" {
			// Only update status and stats if not already completed
			job.Status = "completed"
			job.Progress = 100
			job.UpdateTime = time.Now()
			job.FileCount = len(files)
			updateJobStats(job)
			updateStatusText(job)
		}
		return
	}

	// If not running and not completed, check if it failed
	if job.Status != "queued" && job.Status != "completed" {
		job.Status = "failed"
		job.UpdateTime = time.Now()
		updateStatusText(job)
	}
}

func updateStatusText(job *Job) {
	switch job.Status {
	case "queued":
		job.StatusText = "Queued for processing"
	case "downloading":
		job.StatusText = "Downloading video"
	case "extracting":
		job.StatusText = "Extracting audio"
	case "transcribing":
		job.StatusText = "Transcribing audio"
	case "completed":
		job.StatusText = "Transcription complete"
	case "failed":
		job.StatusText = "Processing failed"
	default:
		job.StatusText = "Processing"
	}
}

func updateJobStats(job *Job) {
	// Calculate duration - only update for running jobs, preserve completed job durations
	if !job.StartTime.IsZero() && job.Status != "completed" && job.Status != "failed" {
		// For running jobs, show elapsed time
		duration := time.Since(job.StartTime)
		job.Duration = formatDuration(duration)
	}

	// Calculate file size
	outputDir := fmt.Sprintf("transcripts/%s", job.VideoID)
	if stat, err := os.Stat(outputDir); err == nil && stat.IsDir() {
		var totalSize int64
		files, _ := os.ReadDir(outputDir)
		job.FileCount = len(files)

		for _, file := range files {
			if fileStat, err := os.Stat(fmt.Sprintf("%s/%s", outputDir, file.Name())); err == nil {
				totalSize += fileStat.Size()
			}
		}
		job.FileSize = formatFileSize(totalSize)
	}
}

func parseJobProgress(job *Job) (string, int) {
	logFiles := []string{
		job.LogFile,
		"transcription.log",
		"nohup.out",
	}

	for _, logFile := range logFiles {
		if status, progress := parseLogFile(logFile); status != "" {
			if title := extractTitleFromLog(logFile); title != "" && job.Title == "Loading..." {
				job.Title = title
			}
			return status, progress
		}
	}

	return "transcribing", job.Progress
}

func parseLogFile(filename string) (string, int) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return "", 0
	}

	content := string(data)
	if strings.Contains(content, "Downloading video") {
		return "downloading", 10
	}
	if strings.Contains(content, "Downloaded:") && strings.Contains(content, "Extracting audio") {
		return "extracting", 30
	}
	if strings.Contains(content, "Audio extracted:") && strings.Contains(content, "Transcribing") {
		return "transcribing", 50
	}
	if strings.Contains(content, "Transcription complete") {
		return "completed", 100
	}

	re := regexp.MustCompile(`(\d+)%`)
	matches := re.FindAllStringSubmatch(content, -1)
	if len(matches) > 0 {
		if percent, err := strconv.Atoi(matches[len(matches)-1][1]); err == nil {
			return "transcribing", 50 + (percent/2)
		}
	}

	return "", 0
}

func extractTitleFromLog(filename string) string {
	cmd := exec.Command("grep", "-o", "Downloaded:.*\\.mp4", filename)
	if output, err := cmd.Output(); err == nil {
		line := strings.TrimSpace(string(output))
		if len(line) > 11 {
			title := line[11:]
			title = strings.TrimSuffix(title, ".mp4")
			title = regexp.MustCompile(`\([^)]*\)`).ReplaceAllString(title, "")
			return strings.TrimSpace(title)
		}
	}
	return ""
}

func extractVideoID(url string) string {
	re := regexp.MustCompile(`(?:youtube\.com/watch\?v=|youtu\.be/|youtube\.com/embed/)([a-zA-Z0-9_-]{11})`)
	matches := re.FindStringSubmatch(url)
	if len(matches) >= 2 {
		return matches[1]
	}
	return "unknown"
}

func generateJobID() string {
	return fmt.Sprintf("job_%d", time.Now().Unix())
}

func formatDuration(d time.Duration) string {
	d = d.Round(time.Second)
	h := d / time.Hour
	d -= h * time.Hour
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second

	if h > 0 {
		return fmt.Sprintf("%02d:%02d:%02d", h, m, s)
	}
	return fmt.Sprintf("%02d:%02d", m, s)
}

func formatFileSize(size int64) string {
	if size < 1024 {
		return fmt.Sprintf("%d B", size)
	}
	if size < 1024*1024 {
		return fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return fmt.Sprintf("%.1f MB", float64(size)/(1024*1024))
}

func calculateBusinessMetrics(data *DashboardData, jobs []Job) {
	// Pricing assumptions for RapidAPI monetization
	basePrice := 2.50  // Base price per job
	premiumMultiplier := 2.0  // Premium tier multiplier

	// Calculate revenue metrics
	data.RevenueToday = float64(data.JobsToday) * basePrice
	data.JobsYesterday = data.JobsToday - 1  // Simulated previous day data

	// Calculate revenue per job with tier detection
	if data.TotalJobs > 0 {
		avgDuration := float64(data.TotalJobs * 5) // Assume 5 min avg
		if avgDuration > 10 { // Longer videos = premium tier
			data.AvgRevenuePerJob = basePrice * premiumMultiplier
		} else {
			data.AvgRevenuePerJob = basePrice
		}
	} else {
		data.AvgRevenuePerJob = basePrice
	}

	// Calculate growth rate (simulated)
	if data.JobsYesterday > 0 {
		data.RevenueGrowth = ((float64(data.JobsToday) - float64(data.JobsYesterday)) / float64(data.JobsYesterday)) * 100
	} else {
		data.RevenueGrowth = 25.0 // Default positive growth for new service
	}
}

// fetchVideoTitle fetches the video title using yt-dlp
func fetchVideoTitle(url string) string {
	// Use yt-dlp to get just the title
	cmd := exec.Command("yt-dlp", "--get-title", "--no-warnings", url)
	output, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(output))
}

// reloadCheckHandler returns file modification timestamp for live reload
func reloadCheckHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	fileMutex.RLock()
	modTime := fileModTime
	fileMutex.RUnlock()

	response := map[string]int64{
		"modified": modTime.UnixMilli(),
	}

	json.NewEncoder(w).Encode(response)
}

func demoAddTranscriptHandler(w http.ResponseWriter, r *http.Request) {
	jobID := strings.TrimPrefix(r.URL.Path, "/demo/add-transcript/")
	if jobID == "" {
		http.Error(w, "Job ID required", http.StatusBadRequest)
		return
	}

	// Check if we have demo data for this job
	if demoData, exists := demoTranscripts[jobID]; exists {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Demo transcript data loaded for job " + jobID,
			"transcript": demoData.Transcript,
			"segments": demoData.Segments,
		})
		return
	}

	http.Error(w, "No demo data available for this job", http.StatusNotFound)
}