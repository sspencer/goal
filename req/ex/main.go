package main

import (
	"fmt"
	"os"

	"github.com/sspencer/goal/req"
)

type Movie struct {
	Title  string `json:"title"`
	Year   int    `json:"year"`
	IMDBID string `json:"imdbID"`
}

type MovieResponse struct {
	Page       int     `json:"page"`
	PerPage    int     `json:"per_page"`
	Total      int     `json:"total"`
	TotalPages int     `json:"total_pages"`
	Data       []Movie `json:"data"`
}

func main() {
	url := "https://jsonmock.hackerrank.com/api/movies/search/?Title=world&page=1"

	r := req.New().CurlHeader()
	resp, err := r.Get(url)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	if !req.IsSuccess(resp.StatusCode) {
		fmt.Printf("movie search failed with HTTP:%d\n", resp.StatusCode)
		os.Exit(2)
	}

	mr := MovieResponse{}
	if err := req.Unmarshal(resp.Body, &mr); err != nil {
		fmt.Println("Could not decode movie response:", err)
		os.Exit(3)
	}

	fmt.Printf("Movie page: %d, per_page: %d, total: %d, total_pages: %d\n", mr.Page, mr.PerPage, mr.Total, mr.TotalPages)
	for _, m := range mr.Data {
		fmt.Printf("  %d: %q\n", m.Year, m.Title)
	}
}
