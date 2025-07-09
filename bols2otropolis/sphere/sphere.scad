sphere_r = 50;
sphere_anchor = "CENTER";
sphere_spin = 0;
sphere_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module sphere_demo(){
	sphere(r=sphere_r, anchor=sphere_anchor, spin=sphere_spin, orient=sphere_orient);
}

sphere_demo(); 