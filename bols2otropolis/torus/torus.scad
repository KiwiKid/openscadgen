torus_r_maj = 50;
torus_r_min = 15;
torus_anchor = "CENTER";
torus_spin = 0;
torus_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module torus_demo(){
	torus(r_maj=torus_r_maj, r_min=torus_r_min, anchor=torus_anchor, spin=torus_spin, orient=torus_orient);
}

torus_demo(); 