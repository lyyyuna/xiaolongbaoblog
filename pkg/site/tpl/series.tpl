
<!DOCTYPE html>
<html>
  <head>
    <meta charset="UTF-8">
    <title>{{.SiteTitle}}: {{.SeriesTitle}} 系列</title>
    
<meta name="viewport" content="width=device-width, initial-scale=1">
<style>
  body {
    padding: 0;
    margin: 0;
    font-size: 100%;
    font-family: serif;
  }
  @media print {
    img {page-break-inside: avoid;}
    div.nosplit {page-break-inside: avoid;}
  }
  img.center {
    display: block;
    margin: 0 auto;
  }
  img.resizable {
    max-width: 100%;
    height: auto;
  }
  .pad {
    padding-top: 1em;
    padding-bottom: 1em;
  }
  a.anchor, a.back, a.footnote {
    color: black !important;
    text-decoration: none !important;
  }
  a.back {
    font-size: 50%;
  }
  @media print {
    a.back {display: none;}
  }
  .header {
    height: 1.25em;
    background-color: #dff;
    margin: 0;
    padding: 0.1em 0.1em 0.2em;
    border-top: 1px solid black;
    border-bottom: 1px solid #8ff;
  }
  .header h3 {
    margin: 0;
    padding: 0 2em;
    display: inline-block;
    padding-right: 2em;
    font-style: italic;
    font-size: 90%;
  }
  .rss {
    float: right;
    padding-top: 0.2em;
    padding-right: 2em;
    display: none;
  }
  .toc {
    margin-top: 2em;
  }
  .toc-title {
    font-family: cursive, serif;
    font-style: italic;
    font-size: 300%;
    line-height: 83%;
  }
  .toc-subtitle {
    display: block;
    margin-bottom: 1em;
    font-size: 83%;
  }
  @media only screen and (max-width: 550px) { .toc-subtitle { display: none; } }
  .header h3 a {
    color: black;
  }
  .header h4 {
    margin: 0;
    padding: 0;
    display: inline-block;
    font-weight: normal;
    font-size: 83%;
  }
  @media only screen and (max-width: 550px) { .header h4 { display: none; } }
  .main {
    padding: 0 2em;
  }
  @media only screen and (max-width: 479px) { .article { font-size: 120%; } }
  .article h1 {
    text-align: center;
    font-size: 200%;
  }
  .copyright {
    font-size: 83%;
  }
  .subtitle {
      font-size: 65%;
  }
  .normal {
    font-size: medium;
    font-weight: normal;
  }
  .when {
    text-align: center;
    font-size: 100%;
    margin: 0;
    padding: 0;
  }
  .when p {
    margin: 0;
    padding: 0;
  }
  .article h2 {
    font-size: 125%;
    padding-top: 0.25em;
  }
  .article h3 {
    font-size: 100%;
  }
  pre {
    margin-left: 4em;
    margin-right: 4em;
  }
  pre, code {
    font-family: monospace;
    font-size: 100%;
  }
  .footer {
    margin-top: 10px;
    font-size: 83%;
    font-family: sans-serif;
  }
  .comments {
    margin-top: 2em;
    background-color: #ffe;
    border-top: 1px solid #aa4;
    border-left: 1px solid #aa4;
    border-right: 1px solid #aa4;
  }
  .comments-header {
    padding: 0 5px 0 5px;
  }
  .comments-header p {
    padding: 0;
    margin: 3px 0 0 0;
  }
  .comments-body {
    padding: 5px 5px 5px 5px;
  }
  #plus-comments {
    border-bottom: 1px dotted #ccc;
  }
  .plus-comment {
    width: 100%;
    font-size: 14px;
    border-top: 1px dotted #ccc;
  }
  .me {
    background-color: #eec;
  }
  .plus-comment ul {
    margin: 0;
    padding: 0;
    list-style: none;
    width: 100%;
    display: inline-block;
  }
  .comment-when {
    color:#999;
    width:auto;
    padding:0 5px;
  }
  .old {
    font-size: 83%;
  }
  .plus-comment ul li {
    display: inline-block;
    vertical-align: top;
    margin-top: 5px;
    margin-bottom: 5px;
    padding: 0;
  }
  .plus-icon {
    width: 45px;
  }
  .plus-img {
    float: left;
    margin: 4px 4px 4px 4px;
    width: 32px;
    height: 32px;
  }
  .plus-comment p {
    margin: 0;
    padding: 0;
  }
  .plus-clear {
    clear: left;
  }
  .toc-when {
    font-size: 83%;
    color: #999;
  }
  .toc {
    list-style: none;
  }
  .toc li {
    margin-bottom: 0.5em;
  }
  .toc-head {
    margin-bottom: 1em !important;
    font-size: 117%;
  }
  .toc-summary {
    margin-left: 2em;
  }
  .favorite {
    font-weight: bold;
  }
  .article p, .article ol {
    line-height: 144%;
  }
  sup, sub {
    vertical-align: baseline;
    position: relative;
    font-size: 83%;
  }
  sup {
    bottom: 1ex;
  }
  sub {
    top: 0.8ex;
  }

  .main {
    position: relative;
    margin: 0 auto;
    padding: 0;
    width: 900px;
  }
  @media only screen and (min-width: 768px) and (max-width: 959px) { .main { width: 708px; } }
  @media only screen and (min-width: 640px) and (max-width: 767px) { .main { width: 580px; } }
  @media only screen and (min-width: 480px) and (max-width: 639px) { .main { width: 420px; } }
  @media only screen and (max-width: 479px) { .main { width: 90%; } }

</style>

</head>
<body>
    
<div class="header">
  <h3><a href="/">{{.SiteTitle}}</a></h3>
  <h4>{{.SiteSubTitle}},
    by <a href="{{.Url}}/about" rel="author">{{.Author}}</a> </h4>
  <a class="rss" href="{{.Url}}/feed.atom">RSS</a>
</div>



<div class="main">
    <ul class="toc">
      <li class="toc-header"><span class="toc-title">{{.SeriesTitle}} 系列</span>
        <span class="toc-subtitle">{{.SiteSubTitle}},
    by <a href="{{.Url}}/about" rel="author">{{.Author}}</a></li>
      
      {{range $blog := .Blogs}}
      <li><a href="{{$blog.Uri}}" class="">{{$blog.Meta.Title}}</a><span class="toc-when">{{$blog.Meta.DateT}}</span>
      
        <div class="toc-summary">
        {{$blog.Meta.Summary}}
        </div>
      </li>
      {{end}}      
    </ul>
</div>

{{.Analytics}}

</body>
</html>
