package keybd_test

var testStrings = map[string]string{
	"shortWord":     "cat",
	"complexWord":   "Pneumonoultramicroscopicsilicovolcanoconiosis",
	"shortSentence": "Good morning!!!",
	"multiLineStringWithTabs": `
function uuidgen
{
	if which uuidgen &>/dev/null; then
		/usr/bin/uuidgen | tr [:upper:] [:lower:]
	else
		cat /proc/sys/kernel/random/uuid
	fi
}
`,
	"multiLineStringWithSpaces": `
function uuidgen
{
    if which uuidgen &>/dev/null; then
        /usr/bin/uuidgen | tr [:upper:] [:lower:]
    else
        cat /proc/sys/kernel/random/uuid
    fi
}
`,
}
