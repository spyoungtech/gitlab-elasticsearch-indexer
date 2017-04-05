require 'net/http'
require 'yaml'

def array(ary)
  "[]string{#{ary.map(&:inspect).join(", ")}}"
end

def bool(value, default)
  value = default if value.nil?
  (!!value).inspect
end

def build_language(name, details)
  out = []

  out << ["Name",         name.inspect]
  out << ["Type",         details['type'].inspect       ] if details['type']
  out << ["Group",        details['group'].inspect      ] if details['group']
  out << ["Color",        details['color'].inspect      ] if details['color']
  out << ["Aliases",      array(details['aliases'])     ] if details['aliases']
  out << ["Extensions",   array(details['extensions'])  ] if details['extensions']
  out << ["Filenames",    array(details['filenames'])   ] if details['filenames']
  out << ["Interpreters", array(details['interpreters'])] if details['interpreters']
  out << ["TmScope",      details['tm_scope'].inspect   ] if details['tm_scope']
  out << ["AceMode",      details['type'].inspect       ] if details['ace_mode']
  out << ["LanguageID",   details['language_id']        ] if details['language_id']

  # Two strange booleans
  out << ["Wrap",         bool(details['wrap'], false)      ]
  out << ["Searchable",   bool(details['searchable'], true) ]

  max_key = out.map {|k,v| k.size }.max
  out = out.map do |k, v|
    "\t\t\t#{k}:#{" " * (max_key - k.size)} #{v},"
  end

  "\t\t#{name.inspect}: &Language{\n#{out.join("\n")}\n\t\t},\n"
end

LANGUAGES_YML = URI.parse("https://raw.githubusercontent.com/github/linguist/v4.7.6/lib/linguist/languages.yml")

languages = YAML.load(Net::HTTP.get(LANGUAGES_YML))

f = File.open("languages.go", "w")
f.puts "package linguist"
f.puts ""
f.puts "var ("
f.puts "\tLanguages = map[string]*Language{"
languages.each {|name, details| f.puts build_language(name, details) }
f.puts "\t}"
f.puts ")"
f.close
