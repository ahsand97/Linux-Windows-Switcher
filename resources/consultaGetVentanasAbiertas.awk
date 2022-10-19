BEGIN {} # Begin section
{
    out = "";
    if ($2 == "-1") {
        next;
    }
    for (i = 5; i <= NF; i++) {
        length(out) > 0 ? out = out" "$i : out = out""$i;
    }
    printf "%d|<>|%s|<>|%s\n", $1, $3, out;
} # Loop section
END {} # End section
