Some models can be fetched from
https://casual-effects.com/data/

Modifications needed:
sponza.mtl:
Change all instances of `map_bump` to `map_Bump`
sponza.obj
Change all instance of `# object` to `o`

Run rexep find&replace in sponza.obj
`vt (.*) (.*) (0\.(?!0000)).*`
replace with
`vt $1 $2 0.0000`

Smoothing groups not supported, remove all `^s [0-9]*$` from sponza.obj
