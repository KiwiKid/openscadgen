octahedron_size = 40;
octahedron_anchor = "CENTER";
octahedron_spin = 0;
octahedron_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module octahedron_demo(){
	octahedron(size=octahedron_size, anchor=octahedron_anchor, spin=octahedron_spin, orient=octahedron_orient);
}

octahedron_demo();
