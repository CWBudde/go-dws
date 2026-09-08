package semantic

// builtinDiagnosticStyle preserves historical wording while the registry owns
// argument counts, parameter types, and results. These descriptions do not
// participate in validation; AST-dependent rules belong in explicit analyzers.
type builtinDiagnosticStyle struct {
	name        string
	arguments   []string
	noArguments bool
	optionalOr  bool
}

var builtinDiagnosticStyles = map[string]builtinDiagnosticStyle{
	"_":               {arguments: []string{"string as argument"}, name: "GetText"},
	"ansicomparestr":  {arguments: []string{"string as first argument", "string as second argument"}},
	"ansicomparetext": {arguments: []string{"string as first argument", "string as second argument"}},
	"ansilowercase":   {arguments: []string{"string as argument"}, name: "LowerCase"},
	"ansiuppercase":   {arguments: []string{"string as argument"}, name: "UpperCase"},
	"asciilowercase":  {arguments: []string{"string as argument"}, name: "LowerCase"},
	"asciiuppercase":  {arguments: []string{"string as argument"}, name: "UpperCase"},
	"bytesizetostr":   {arguments: []string{"integer as argument"}},
	"comparestr":      {arguments: []string{"string as first argument", "string as second argument"}},
	"comparetext":     {arguments: []string{"string as first argument", "string as second argument"}},
	"deleteleft":      {arguments: []string{"string as first argument", "integer as second argument"}, name: "StrDeleteLeft"},
	"deleteright":     {arguments: []string{"string as first argument", "integer as second argument"}, name: "StrDeleteRight"},
	"dupestring":      {arguments: []string{"string as first argument", "integer as second argument"}, name: "StringOfString"},
	"factorial":       {arguments: []string{"Integer argument"}},
	"frac":            {arguments: []string{"Float or Integer"}},
	"gettext":         {arguments: []string{"string as argument"}},
	"infinity":        {noArguments: true},
	"int":             {arguments: []string{"Float or Integer"}},
	"isdelimiter":     {arguments: []string{"string as first argument", "string as second argument", "integer as third argument"}},
	"isfinite":        {arguments: []string{"Float or Integer"}},
	"isinfinite":      {arguments: []string{"Float or Integer"}},
	"isprime":         {arguments: []string{"Integer argument"}},
	"lastdelimiter":   {arguments: []string{"string as first argument", "string as second argument"}},
	"leastfactor":     {arguments: []string{"Integer argument"}},
	"log10":           {arguments: []string{"Float or Integer"}},
	"lowercase":       {arguments: []string{"string as argument"}},
	"nan":             {noArguments: true},
	"normalize":       {optionalOr: true, arguments: []string{"string as first argument", "string as second argument"}, name: "NormalizeString"},
	"normalizestring": {optionalOr: true, arguments: []string{"string as first argument", "string as second argument"}},
	"odd":             {arguments: []string{"Integer"}},
	"padleft":         {optionalOr: true, arguments: []string{"string as first argument", "integer as second argument", "string as third argument"}},
	"padright":        {optionalOr: true, arguments: []string{"string as first argument", "integer as second argument", "string as third argument"}},
	"pi":              {noArguments: true},
	"popcount":        {arguments: []string{"Integer argument"}},
	"posex":           {arguments: []string{"string as first argument", "string as second argument", "integer as third argument"}},
	"quotedstr":       {optionalOr: true, arguments: []string{"string as first argument", "string as second argument"}},
	"random":          {noArguments: true},
	"randomint":       {arguments: []string{"Integer argument"}},
	"randseed":        {noArguments: true},
	"reversestring":   {arguments: []string{"string as argument"}},
	"revpos":          {arguments: []string{"string as first argument", "string as second argument"}},
	"rightstr":        {arguments: []string{"string as first argument", "integer as second argument"}},
	"sametext":        {arguments: []string{"string as first argument", "string as second argument"}},
	"strafter":        {arguments: []string{"string as first argument", "string as second argument"}},
	"strafterlast":    {arguments: []string{"string as first argument", "string as second argument"}},
	"strbefore":       {arguments: []string{"string as first argument", "string as second argument"}},
	"strbeforelast":   {arguments: []string{"string as first argument", "string as second argument"}},
	"strbeginswith":   {arguments: []string{"string as first argument", "string as second argument"}},
	"strbetween":      {arguments: []string{"string as first argument", "string as second argument", "string as third argument"}},
	"strcontains":     {arguments: []string{"string as first argument", "string as second argument"}},
	"strdeleteleft":   {arguments: []string{"string as first argument", "integer as second argument"}},
	"strdeleteright":  {arguments: []string{"string as first argument", "integer as second argument"}},
	"strendswith":     {arguments: []string{"string as first argument", "string as second argument"}},
	"stringofstring":  {arguments: []string{"string as first argument", "integer as second argument"}},
	"stripaccents":    {arguments: []string{"string as argument"}},
	"strisascii":      {arguments: []string{"string as argument"}},
	"strmatches":      {arguments: []string{"string as first argument", "string as second argument"}},
	"unsigned32":      {arguments: []string{"Integer argument"}},
	"uppercase":       {arguments: []string{"string as argument"}},
}
