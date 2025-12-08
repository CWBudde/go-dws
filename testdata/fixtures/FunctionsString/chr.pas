// Test Chr() function with UTF-8 encoding
// NOTE: go-dws uses UTF-8, not UTF-16 like DWScript
// So Chr() returns single Unicode characters, not surrogate pairs

// Basic ASCII
if Chr($41)<>'A' then PrintLn('bug A');

// BMP (Basic Multilingual Plane)
if Chr($263A)<>'☺' then PrintLn('bug smiling face');

// Supplementary planes (U+10000 and above)
// In UTF-8, these are single characters with Length=1
// In UTF-16 (DWScript), these would be surrogate pairs with Length=2
if Length(Chr($10000))<>1 then PrintLn('bug U+10000 length');
if Length(Chr($103FF))<>1 then PrintLn('bug U+103FF length');
if Length(Chr($10FC00))<>1 then PrintLn('bug U+10FC00 length');
if Length(Chr($10FC01))<>1 then PrintLn('bug U+10FC01 length');
if Length(Chr($10FFFF))<>1 then PrintLn('bug U+10FFFF length');

// Verify the characters are valid (can be used in strings)
var s := Chr($10000) + Chr($1F600);
if Length(s)<>2 then PrintLn('bug string concatenation');

PrintLn(Chr($42));