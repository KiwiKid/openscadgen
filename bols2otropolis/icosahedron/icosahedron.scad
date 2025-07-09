icosahedron_size = 40;
icosahedron_anchor = "CENTER";
icosahedron_spin = 0;
icosahedron_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module icosahedron_demo(){
	icosahedron(size=icosahedron_size, anchor=icosahedron_anchor, spin=icosahedron_spin, orient=icosahedron_orient);
}

icosahedron_demo(); 