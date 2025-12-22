//go:build darwin && appstore

package services

/*
#cgo CFLAGS: -x objective-c
#cgo LDFLAGS: -framework AVFoundation -framework CoreMedia -framework CoreVideo -framework Foundation

#import <AVFoundation/AVFoundation.h>
#import <Foundation/Foundation.h>
#include <stdlib.h>

// Forward declaration for Go callback
extern void goProgressCallback(void* callback, float progress);

// Convert video using AVFoundation
static int convertVideoWithAVFoundation(const char* inputPath, const char* outputPath, const char* preset, void* progressCallback) {
    @autoreleasepool {
        NSString* input = [NSString stringWithUTF8String:inputPath];
        NSString* output = [NSString stringWithUTF8String:outputPath];
        NSString* presetName = [NSString stringWithUTF8String:preset];

        NSURL* inputURL = [NSURL fileURLWithPath:input];
        NSURL* outputURL = [NSURL fileURLWithPath:output];

        // Remove existing output file
        [[NSFileManager defaultManager] removeItemAtURL:outputURL error:nil];

        AVAsset* asset = [AVAsset assetWithURL:inputURL];
        if (!asset) {
            return -1;
        }

        // Determine output file type based on extension
        NSString* extension = [[output pathExtension] lowercaseString];
        AVFileType fileType = AVFileTypeMPEG4;
        if ([extension isEqualToString:@"mov"]) {
            fileType = AVFileTypeQuickTimeMovie;
        } else if ([extension isEqualToString:@"m4v"]) {
            fileType = AVFileTypeAppleM4V;
        }

        AVAssetExportSession* exportSession = [[AVAssetExportSession alloc] initWithAsset:asset presetName:presetName];
        if (!exportSession) {
            return -2;
        }

        exportSession.outputURL = outputURL;
        exportSession.outputFileType = fileType;
        exportSession.shouldOptimizeForNetworkUse = YES;

        // Create a semaphore to wait for completion
        dispatch_semaphore_t semaphore = dispatch_semaphore_create(0);
        __block int result = 0;

        // Start progress monitoring in background
        void* cbPtr = progressCallback;
        dispatch_async(dispatch_get_global_queue(DISPATCH_QUEUE_PRIORITY_DEFAULT, 0), ^{
            while (exportSession.status == AVAssetExportSessionStatusExporting ||
                   exportSession.status == AVAssetExportSessionStatusWaiting) {
                if (cbPtr != NULL) {
                    goProgressCallback(cbPtr, exportSession.progress * 100.0);
                }
                [NSThread sleepForTimeInterval:0.1];
            }
        });

        [exportSession exportAsynchronouslyWithCompletionHandler:^{
            switch (exportSession.status) {
                case AVAssetExportSessionStatusCompleted:
                    result = 0;
                    break;
                case AVAssetExportSessionStatusFailed:
                    NSLog(@"AVFoundation export failed: %@", exportSession.error);
                    result = -3;
                    break;
                case AVAssetExportSessionStatusCancelled:
                    result = -4;
                    break;
                default:
                    result = -5;
                    break;
            }
            dispatch_semaphore_signal(semaphore);
        }];

        dispatch_semaphore_wait(semaphore, DISPATCH_TIME_FOREVER);
        return result;
    }
}

*/
import "C"

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	"unsafe"

	"zenfile/internal/logger"
	"zenfile/internal/models"
)

// Callback registry for progress updates
var (
	callbackMutex   sync.RWMutex
	callbackCounter uintptr
	callbackMap     = make(map[uintptr]func(float64))
)

//export goProgressCallback
func goProgressCallback(callbackPtr unsafe.Pointer, progress C.float) {
	ptr := uintptr(callbackPtr)
	callbackMutex.RLock()
	callback, ok := callbackMap[ptr]
	callbackMutex.RUnlock()
	if ok && callback != nil {
		callback(float64(progress))
	}
}

func registerCallback(fn func(float64)) uintptr {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	callbackCounter++
	callbackMap[callbackCounter] = fn
	return callbackCounter
}

func unregisterCallback(ptr uintptr) {
	callbackMutex.Lock()
	defer callbackMutex.Unlock()
	delete(callbackMap, ptr)
}

// avfVideoConverter handles video file conversion using AVFoundation
type avfVideoConverter struct {
	log *logger.ComponentLogger
}

// NewVideoConverter creates a new video converter using AVFoundation (App Store build)
func NewVideoConverter(_ interface{}, log *logger.Logger) Converter {
	return &avfVideoConverter{
		log: log.WithComponent("avf-video-converter"),
	}
}

// Convert converts a video file using AVFoundation
func (c *avfVideoConverter) Convert(job models.ConversionJob, progressCallback func(progress float64)) (*models.ConversionResult, error) {
	c.log.Info("Starting AVFoundation video conversion: %s -> %s", job.InputPath, job.OutputPath)
	startTime := time.Now()

	result := &models.ConversionResult{
		InputPath:  job.InputPath,
		OutputPath: job.OutputPath,
	}

	// Validate input file exists
	if _, err := os.Stat(job.InputPath); os.IsNotExist(err) {
		result.ErrorMessage = fmt.Sprintf("Input file not found: %s", job.InputPath)
		c.log.Error(result.ErrorMessage)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	// Check output directory exists
	outputDir := filepath.Dir(job.OutputPath)
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		result.ErrorMessage = fmt.Sprintf("Failed to create output directory: %v", err)
		c.log.Error(result.ErrorMessage)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	// Check if output file already exists
	if !job.OverwriteOutput {
		if _, err := os.Stat(job.OutputPath); err == nil {
			result.ErrorMessage = fmt.Sprintf("Output file already exists: %s", job.OutputPath)
			c.log.Error(result.ErrorMessage)
			return result, fmt.Errorf(result.ErrorMessage)
		}
	}

	// Get output format
	outputFormat := strings.TrimPrefix(strings.ToLower(filepath.Ext(job.OutputPath)), ".")

	// Check if format is supported by AVFoundation
	if !c.isAVFoundationFormat(outputFormat) {
		result.ErrorMessage = fmt.Sprintf("Format %s is not supported by AVFoundation. Supported formats: mp4, mov, m4v", outputFormat)
		c.log.Error(result.ErrorMessage)
		return result, fmt.Errorf(result.ErrorMessage)
	}

	// Determine the best preset
	preset := c.getPresetForFormat(outputFormat)

	// Register progress callback
	var callbackPtr uintptr
	if progressCallback != nil {
		callbackPtr = registerCallback(progressCallback)
		defer unregisterCallback(callbackPtr)
	}

	// Convert using AVFoundation
	inputCStr := C.CString(job.InputPath)
	outputCStr := C.CString(job.OutputPath)
	presetCStr := C.CString(preset)
	defer C.free(unsafe.Pointer(inputCStr))
	defer C.free(unsafe.Pointer(outputCStr))
	defer C.free(unsafe.Pointer(presetCStr))

	var cbPtr unsafe.Pointer
	if callbackPtr != 0 {
		cbPtr = unsafe.Pointer(callbackPtr)
	}

	ret := C.convertVideoWithAVFoundation(inputCStr, outputCStr, presetCStr, cbPtr)

	if ret != 0 {
		var errMsg string
		switch ret {
		case -1:
			errMsg = "Failed to load input asset"
		case -2:
			errMsg = "Failed to create export session"
		case -3:
			errMsg = "Export failed"
		case -4:
			errMsg = "Export was cancelled"
		default:
			errMsg = fmt.Sprintf("Unknown error: %d", ret)
		}
		result.ErrorMessage = errMsg
		c.log.Error("AVFoundation conversion failed: %s", errMsg)
		return result, fmt.Errorf(errMsg)
	}

	// Get output file size
	if stat, err := os.Stat(job.OutputPath); err == nil {
		result.OutputSize = stat.Size()
	}

	if progressCallback != nil {
		progressCallback(100)
	}

	result.Success = true
	result.Duration = time.Since(startTime).Milliseconds()

	c.log.Info("AVFoundation video conversion completed in %dms: %s", result.Duration, job.OutputPath)
	return result, nil
}

// isAVFoundationFormat checks if the format is supported by AVFoundation
func (c *avfVideoConverter) isAVFoundationFormat(format string) bool {
	supported := map[string]bool{
		"mp4": true,
		"mov": true,
		"m4v": true,
	}
	return supported[format]
}

// getPresetForFormat returns the appropriate AVAssetExportSession preset
func (c *avfVideoConverter) getPresetForFormat(format string) string {
	// Use highest quality preset - AVFoundation handles codec selection
	return "AVAssetExportPresetHighestQuality"
}

// SupportedInputFormats returns the list of supported input video formats
func (c *avfVideoConverter) SupportedInputFormats() []string {
	// AVFoundation natively supports QuickTime-compatible formats
	// Note: webm, mkv, avi are NOT supported by AVFoundation without additional plugins
	return []string{"mp4", "mov", "m4v", "3gp"}
}

// SupportedOutputFormats returns the list of supported output formats for video
// AVFoundation has more limited output format support than FFmpeg
func (c *avfVideoConverter) SupportedOutputFormats(inputFormat string) []string {
	// AVFoundation primarily supports Apple formats for output
	return []string{"mp4", "mov", "m4v"}
}

// CanConvert checks if conversion is possible between formats
func (c *avfVideoConverter) CanConvert(inputFormat, outputFormat string) bool {
	inputFormat = strings.ToLower(strings.TrimPrefix(inputFormat, "."))
	outputFormat = strings.ToLower(strings.TrimPrefix(outputFormat, "."))

	// Check input is supported
	inputSupported := false
	for _, f := range c.SupportedInputFormats() {
		if f == inputFormat {
			inputSupported = true
			break
		}
	}
	if !inputSupported {
		return false
	}

	// Check output is supported
	return c.isAVFoundationFormat(outputFormat)
}
