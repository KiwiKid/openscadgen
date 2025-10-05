cone_h = 60;
cone_r1 = 40;
cone_r2 = 0;
cone_center = true;
cone_anchor = "CENTER";
cone_spin = 0;
cone_orient = undef;

include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

module cone_demo(){
	cyl(h=cone_h, r1=cone_r1, r2=cone_r2, center=cone_center, anchor=cone_anchor, spin=cone_spin, orient=cone_orient);
}

cone_demo(); 