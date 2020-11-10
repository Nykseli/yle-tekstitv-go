package main

/*
#cgo CFLAGS: -I./lib
#cgo LDFLAGS: -L./lib -ltekstitv -Wl,-rpath=./lib -lcurl
#include "tekstitv.h"
#include <stdlib.h>
// Helper for finding the row index pointer
// This is a bit messy in go so here is a simple function
// for making life easier
html_row* row_index_helper(html_row* row, int idx) {
	return &row[idx];
}
*/
import "C"
import (
	"fmt"
	"os"
	"strconv"
	"unsafe"
)

const (
	// TypeHTMLText represents the HTMLText
	TypeHTMLText = C.HTML_TEXT
	// TypeHTMLLink represents the HTMLLink
	TypeHTMLLink = C.HTML_LINK
)

// HTMLText is struct html_text
type HTMLText struct {
	text string
}

// HTMLLink is struct html_link
type HTMLLink struct {
	url       string
	innerText string
}

// HTMLItem is struct html_item
type HTMLItem struct {
	itemType int
	item     interface{}
}

// HTMLRow is struct html_row
type HTMLRow struct {
	items []HTMLItem
}

// HTMLParser is struct html_parser
type HTMLParser struct {
	title            HTMLText
	topNavigation    []HTMLItem
	bottomNavigation []HTMLLink
	subPages         HTMLRow
	middle           []HTMLRow
	loadError        bool
}

func cHTMLText(text *C.html_text) HTMLText {
	// &text.text[0] gets the pointer to the first char
	// turning char[] into char*
	gostr := C.GoString(&text.text[0])
	return HTMLText{text: gostr}
}

func cHTMLLink(link *C.html_link) HTMLLink {
	// &link.text[0] gets the pointer to the first char
	// turning char[] into char*
	urlStr := C.GoString(&link.url.text[0])
	innerStr := C.GoString(&link.inner_text.text[0])
	return HTMLLink{url: urlStr, innerText: innerStr}
}

// ItemAsText turn a HTMLItem into HTMLText
func ItemAsText(item HTMLItem) HTMLText {
	return item.item.(HTMLText)
}

// ItemAsLink turn a HTMLItem into HTMLLink
func ItemAsLink(item HTMLItem) HTMLLink {
	return item.item.(HTMLLink)
}

func cHTMLItem(htmlItem *C.html_item) HTMLItem {
	// It should be type but it's changed to i_type in ./build_third_party.sh
	// because type is reserved word in go
	iType := int(htmlItem.i_type)
	var tmpItem interface{}
	switch iType {
	case TypeHTMLText:
		tmpItem = cHTMLText((*C.html_text)(unsafe.Pointer(&htmlItem.item[0])))
		break
	case TypeHTMLLink:
		tmpItem = cHTMLLink((*C.html_link)(unsafe.Pointer(&htmlItem.item[0])))
		break
		// TODO: default error
	}

	return HTMLItem{itemType: iType, item: tmpItem}
}
func cHTMLRow(row *C.html_row) HTMLRow {
	rowSize := int(row.size)
	tmpItems := make([]HTMLItem, rowSize, rowSize)
	for i := 0; i < rowSize; i++ {
		tmpItems[i] = cHTMLItem(&row.items[i])
	}

	return HTMLRow{items: tmpItems}
}
func cHTMLParser(parser *C.html_parser) HTMLParser {
	tmpTitle := cHTMLText(&parser.title)
	tmpSubPages := cHTMLRow(&parser.sub_pages)
	tmpLoadError := bool(parser.curl_load_error)

	tmpTopNavigation := make([]HTMLItem, C.TOP_NAVIGATION_SIZE, C.TOP_NAVIGATION_SIZE)
	for i := 0; i < C.TOP_NAVIGATION_SIZE; i++ {
		tmpTopNavigation[i] = cHTMLItem(&parser.top_navigation[i])
	}

	tmpBottomNavigation := make([]HTMLLink, C.BOTTOM_NAVIGATION_SIZE, C.BOTTOM_NAVIGATION_SIZE)
	for i := 0; i < C.BOTTOM_NAVIGATION_SIZE; i++ {
		tmpBottomNavigation[i] = cHTMLLink(&parser.bottom_navigation[i])
	}

	tmpMiddle := make([]HTMLRow, parser.middle_rows, parser.middle_rows)
	for i := 0; i < int(parser.middle_rows); i++ {
		// row_index_helper is so much cleaner than the
		// unsafe stuff that's commented out
		rowItem := C.row_index_helper(parser.middle, C.int(i))
		tmpMiddle[i] = cHTMLRow(rowItem)
		// tmpMiddle[i] = cHTMLRow((*C.html_row)(unsafe.Pointer(uintptr(unsafe.Pointer(parser.middle)) + 6*unsafe.Sizeof(parser.middle))))
	}

	return HTMLParser{
		title:            tmpTitle,
		subPages:         tmpSubPages,
		loadError:        tmpLoadError,
		topNavigation:    tmpTopNavigation,
		bottomNavigation: tmpBottomNavigation,
		middle:           tmpMiddle,
	}
}

func printTitle(parser *HTMLParser) {
	fmt.Printf("\n  %s\n", parser.title.text)
}

func printMiddle(parser *HTMLParser) {
	lastType := TypeHTMLText
	for i := 0; i < len(parser.middle); i++ {
		fmt.Printf("  ")
		for j := 0; j < len(parser.middle[i].items); j++ {
			htmlItem := parser.middle[i].items[j]
			switch htmlItem.itemType {
			case TypeHTMLText:
				fmt.Printf("%s", ItemAsText(htmlItem).text)
				lastType = TypeHTMLText
				break
			case TypeHTMLLink:
				if lastType == TypeHTMLLink {
					fmt.Printf("-")
				}
				fmt.Printf("%s", ItemAsLink(htmlItem).innerText)
				lastType = TypeHTMLLink
				break
			}
		}
		fmt.Printf("\n")
	}

	fmt.Printf("  ")
	for i := 0; i < len(parser.subPages.items); i++ {
		htmlItem := parser.subPages.items[i]
		switch htmlItem.itemType {
		case TypeHTMLText:
			fmt.Printf("%s", ItemAsText(htmlItem).text)
			break
		case TypeHTMLLink:
			fmt.Printf("%s", ItemAsLink(htmlItem).innerText)
			break
		}
	}

	fmt.Printf("\n\n")
}

func main() {
	page := 100

	args := os.Args

	if len(args) > 1 {
		page, _ = strconv.Atoi(args[1])
	}

	var parserStruct C.html_parser
	C.init_html_parser(&parserStruct)
	C.link_from_ints(&parserStruct, C.int(page), C.int(1))
	C.load_page(&parserStruct)
	C.parse_html(&parserStruct)

	testH := cHTMLParser(&parserStruct)
	if testH.loadError {
		fmt.Printf("Failed load the page\n")
	} else {
		printTitle(&testH)
		printMiddle(&testH)
	}

	C.free_html_parser(&parserStruct)
}
