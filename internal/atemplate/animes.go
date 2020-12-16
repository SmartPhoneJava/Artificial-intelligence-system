package atemplate

import (
	"fmt"
	"html/template"
)

func ShowAnimes(s interface{}) template.HTML {
	return template.HTML(fmt.Sprint(s))
}

func ShowBranch(branch []string) template.HTML {
	return template.HTML(
		`
	<div class="b-breadcrumbs">
		{{ range .Branch }}
		<span itemscope="" itemtype="http://data-vocabulary.org/Breadcrumb"><a class="b-link"
				href="https://shikimori.one/animes" itemprop="url" title="Аниме"><span
					itemprop="title"> {{ . }}</span></a>
		</span>
		{{ end }}
	</div>
	`)
}
