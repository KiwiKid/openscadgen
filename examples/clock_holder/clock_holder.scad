
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


baseSize = [200, 30, 10];
clockSize = [20, 25, 180];
 
module clock_holder(){
	difference(){
		cuboid(baseSize, rounding1=-1);
        up(10)
        rotate([0,90,0])
		cuboid(clockSize, rounding=1, edges="X");
		
	}
}

clock_holder();
