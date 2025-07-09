dodecahedron_size = 35;
dodecahedron_anchor = "CENTER";
dodecahedron_spin = 0;
dodecahedron_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module dodecahedron_demo(){
	dodecahedron(size=dodecahedron_size, anchor=dodecahedron_anchor, spin=dodecahedron_spin, orient=dodecahedron_orient);
}

dodecahedron_demo(); 