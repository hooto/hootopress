<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1, shrink-to-fit=no">
  <title>{{if .__html_head_title__}}{{.__html_head_title__}} | {{end}}{{SysConfig "frontend_html_head_subtitle"}}</title>
  <link rel="stylesheet" href="{{HttpSrvBasePath "hp/~/bulma/v0/css/bulma.css"}}?v={{.sys_version_sign}}" type="text/css">
  <link rel="stylesheet" href="{{HttpSrvBasePath "hp/~/hp/css/main.v2.css"}}?v={{.sys_version_sign}}" type="text/css">
  {{if SysConfig "frontend_header_site_icon_url"}}
  <link rel="shortcut icon" type="image/x-icon" href="{{SysConfig "frontend_header_site_icon_url"}}?v={{.sys_version_sign}}">
  {{else}}
  <link rel="shortcut icon" type="image/x-icon" href="{{HttpSrvBasePath "hp/~/hp/img/ap.ico"}}?v={{.sys_version_sign}}">
  {{end}}
  <meta name="keywords" content="{{SysConfig "frontend_html_head_meta_keywords"}}">
  <meta name="description" content="{{SysConfig "frontend_html_head_meta_description"}}">
  <script src="{{HttpSrvBasePath "hp/~/lessui/js/sea.js"}}?v={{.sys_version_sign}}"></script>
  <script src="{{HttpSrvBasePath "hp/~/hp/js/main.v2.js"}}?v={{.sys_version_sign}}"></script>
  <script type="text/javascript">
    window._basepath = {{HttpSrvBasePath "hp/"}};
    window._sys_version_sign = {{.sys_version_sign}};
    window.onload_hooks = [];
  </script>
</head>
