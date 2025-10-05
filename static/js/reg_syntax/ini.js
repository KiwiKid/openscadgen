editAreaLoader.load_syntax["ini"] = {
	'DISPLAY_NAME' : 'INI/TOML'
	,'COMMENT_SINGLE' : {1 : '#'}
	,'QUOTEMARKS' : ['"', "'"]
	,'KEYWORD_CASE_SENSITIVE' : false
	,'KEYWORDS' : {
		'keywords' : [
			'true', 'false', 'yes', 'no', 'on', 'off'
		]
		,'sections' : [
			'[', ']'
		]
	}
	,'OPERATORS' : [
		'=', '+=', '-=', '*=', '/=', '%=', '&=', '|=', '^=', '<<=', '>>='
	]
	,'DELIMITERS' : [
		'(', ')', '[', ']', '{', '}'
	]
	,'REGEXPS' : {
		'section' : {
			'search' : '^\\s*\\[.*\\]\\s*$'
			,'class' : 'ini_section'
			,'modifiers' : 'm'
		}
		,'key' : {
			'search' : '^\\s*([a-zA-Z_][a-zA-Z0-9_]*)\\s*='
			,'class' : 'ini_key'
			,'modifiers' : 'm'
		}
		,'number' : {
			'search' : '\\b\\d+(\\.\\d+)?\\b'
			,'class' : 'ini_number'
		}
	}
	,'STYLES' : {
		'COMMENT_SINGLE' : 'color: #808080;'
		,'COMMENT_MULTI' : 'color: #808080;'
		,'KEYWORDS' : {
			'keywords' : 'color: #0000ff;'
			,'sections' : 'color: #800080; font-weight: bold;'
		}
		,'OPERATORS' : 'color: #ff0000;'
		,'DELIMITERS' : 'color: #ff0000;'
		,'REGEXPS' : {
			'section' : 'color: #800080; font-weight: bold;'
			,'key' : 'color: #000080; font-weight: bold;'
			,'number' : 'color: #008000;'
		}
	}
};