package imaging

import (
	"bytes"
	"context"
	"encoding/base64"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

// gopher was created by Takuya Ueda (https://twitter.com/tenntenn). Licensed under the Creative Commons 3.0 Attributions license.
// https://creativecommons.org/licenses/by/3.0/deed.ja
const gopher = `<?xml version="1.0" encoding="utf-8"?>
<!-- Generator: Adobe Illustrator 15.0.0, SVG Export Plug-In . SVG Version: 6.00 Build 0)  -->
<!DOCTYPE svg PUBLIC "-//W3C//DTD SVG 1.1//EN" "http://www.w3.org/Graphics/SVG/1.1/DTD/svg11.dtd">
<svg version="1.1" id="レイヤー_1" xmlns="http://www.w3.org/2000/svg" xmlns:xlink="http://www.w3.org/1999/xlink" x="0px"
	 y="0px" width="401.98px" height="559.472px" viewBox="0 0 401.98 559.472" enable-background="new 0 0 401.98 559.472"
	 xml:space="preserve">
<path fill-rule="evenodd" clip-rule="evenodd" fill="#F6D2A2" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M10.634,300.493c0.764,15.751,16.499,8.463,23.626,3.539c6.765-4.675,8.743-0.789,9.337-10.015
	c0.389-6.064,1.088-12.128,0.744-18.216c-10.23-0.927-21.357,1.509-29.744,7.602C10.277,286.542,2.177,296.561,10.634,300.493"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#C6B198" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M10.634,300.493c2.29-0.852,4.717-1.457,6.271-3.528"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#6AD7E5" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M46.997,112.853C-13.3,95.897,31.536,19.189,79.956,50.74L46.997,112.853z"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#6AD7E5" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M314.895,44.984c47.727-33.523,90.856,42.111,35.388,61.141L314.895,44.984z"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#F6D2A2" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M325.161,494.343c12.123,7.501,34.282,30.182,16.096,41.18c-17.474,15.999-27.254-17.561-42.591-22.211
	C305.271,504.342,313.643,496.163,325.161,494.343z"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="none" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M341.257,535.522c-2.696-5.361-3.601-11.618-8.102-15.939"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#F6D2A2" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M108.579,519.975c-14.229,2.202-22.238,15.039-34.1,21.558c-11.178,6.665-15.454-2.134-16.461-3.92
	c-1.752-0.799-1.605,0.744-4.309-1.979c-10.362-16.354,10.797-28.308,21.815-36.432C90.87,496.1,100.487,509.404,108.579,519.975z"
	/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="none" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M58.019,537.612c0.542-6.233,5.484-10.407,7.838-15.677"/>
<path fill-rule="evenodd" clip-rule="evenodd" d="M49.513,91.667c-7.955-4.208-13.791-9.923-8.925-19.124
	c4.505-8.518,12.874-7.593,20.83-3.385L49.513,91.667z"/>
<path fill-rule="evenodd" clip-rule="evenodd" d="M337.716,83.667c7.955-4.208,13.791-9.923,8.925-19.124
	c-4.505-8.518-12.874-7.593-20.83-3.385L337.716,83.667z"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#F6D2A2" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M392.475,298.493c-0.764,15.751-16.499,8.463-23.626,3.539c-6.765-4.675-8.743-0.789-9.337-10.015
	c-0.389-6.064-1.088-12.128-0.744-18.216c10.23-0.927,21.357,1.509,29.744,7.602C392.831,284.542,400.932,294.561,392.475,298.493"
	/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#C6B198" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M392.475,298.493c-2.29-0.852-4.717-1.457-6.271-3.528"/>
<g>
	<path fill-rule="evenodd" clip-rule="evenodd" fill="#6AD7E5" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
		M195.512,13.124c60.365,0,116.953,8.633,146.452,66.629c26.478,65.006,17.062,135.104,21.1,203.806
		c3.468,58.992,11.157,127.145-16.21,181.812c-28.79,57.514-100.73,71.982-160,69.863c-46.555-1.666-102.794-16.854-129.069-59.389
		c-30.826-49.9-16.232-124.098-13.993-179.622c2.652-65.771-17.815-131.742,3.792-196.101
		C69.999,33.359,130.451,18.271,195.512,13.124"/>
</g>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="2.9081" stroke-linecap="round" d="
	M206.169,94.16c10.838,63.003,113.822,46.345,99.03-17.197C291.935,19.983,202.567,35.755,206.169,94.16"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="2.8214" stroke-linecap="round" d="
	M83.103,104.35c14.047,54.85,101.864,40.807,98.554-14.213C177.691,24.242,69.673,36.957,83.103,104.35"/>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M218.594,169.762c0.046,8.191,1.861,17.387,0.312,26.101c-2.091,3.952-6.193,4.37-9.729,5.967c-4.89-0.767-9.002-3.978-10.963-8.552
	c-1.255-9.946,0.468-19.576,0.785-29.526L218.594,169.762z"/>
<g>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" cx="107.324" cy="95.404" rx="14.829" ry="16.062"/>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" cx="114.069" cy="99.029" rx="3.496" ry="4.082"/>
</g>
<g>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" cx="231.571" cy="91.404" rx="14.582" ry="16.062"/>
	<ellipse fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" cx="238.204" cy="95.029" rx="3.438" ry="4.082"/>
</g>
<path fill-rule="evenodd" clip-rule="evenodd" fill="#FFFFFF" stroke="#000000" stroke-width="3" stroke-linecap="round" d="
	M176.217,168.87c-6.47,15.68,3.608,47.035,21.163,23.908c-1.255-9.946,0.468-19.576,0.785-29.526L176.217,168.87z"/>
<g>
	<path fill-rule="evenodd" clip-rule="evenodd" fill="#F6D2A2" stroke="#231F20" stroke-width="3" stroke-linecap="round" d="
		M178.431,138.673c-12.059,1.028-21.916,15.366-15.646,26.709c8.303,15.024,26.836-1.329,38.379,0.203
		c13.285,0.272,24.17,14.047,34.84,2.49c11.867-12.854-5.109-25.373-18.377-30.97L178.431,138.673z"/>
	<path fill-rule="evenodd" clip-rule="evenodd" d="M176.913,138.045c-0.893-20.891,38.938-23.503,43.642-6.016
		C225.247,149.475,178.874,153.527,176.913,138.045C175.348,125.682,176.913,138.045,176.913,138.045z"/>
</g>
</svg>`

// from https://gifer.com/en/N9v8
const base64gif = `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`

func TestConvertToPNG(t *testing.T) {
	t.Parallel()

	im := os.Getenv("TRAQ_IMAGEMAGICK")
	if len(im) == 0 {
		t.SkipNow()
	}

	t.Run("unavailable", func(t *testing.T) {
		t.Parallel()

		_, err := ConvertToPNG(context.TODO(), "", bytes.NewBufferString(""), 100, 200)
		assert.Error(t, err)
	})

	t.Run("Broken svg", func(t *testing.T) {
		t.Parallel()

		broken := `<?xml version="1.0" enc`
		_, err := ConvertToPNG(context.TODO(), im, bytes.NewBufferString(broken), 100, 200)
		assert.Error(t, err)
	})

	t.Run("Valid svg", func(t *testing.T) {
		t.Parallel()

		_, err := ConvertToPNG(context.TODO(), im, bytes.NewBufferString(gopher), 100, 100)
		assert.NoError(t, err)
	})

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		_, err := ConvertToPNG(context.TODO(), im, bytes.NewBufferString(gopher), -100, 100)
		assert.Error(t, err)
	})
}

func TestResizeAnimationGIF(t *testing.T) {
	t.Parallel()

	im := os.Getenv("TRAQ_IMAGEMAGICK")
	if len(im) == 0 {
		t.SkipNow()
	}

	gif, _ := base64.RawStdEncoding.DecodeString(base64gif)

	t.Run("unavailable", func(t *testing.T) {
		t.Parallel()

		_, err := ResizeAnimationGIF(context.TODO(), "", bytes.NewBufferString(""), 100, 200, false)
		assert.Error(t, err)
	})

	t.Run("valid gif", func(t *testing.T) {
		t.Parallel()

		_, err := ResizeAnimationGIF(context.TODO(), im, bytes.NewReader(gif), 50, 50, false)
		assert.NoError(t, err)
	})

	t.Run("not gif", func(t *testing.T) {
		t.Parallel()

		_, err := ResizeAnimationGIF(context.TODO(), im, io.LimitReader(bytes.NewReader(gif), 10), 100, 100, true)
		assert.Error(t, err)
	})

	t.Run("invalid args", func(t *testing.T) {
		t.Parallel()

		_, err := ResizeAnimationGIF(context.TODO(), im, bytes.NewBufferString(gopher), -100, 100, false)
		assert.Error(t, err)
	})
}
