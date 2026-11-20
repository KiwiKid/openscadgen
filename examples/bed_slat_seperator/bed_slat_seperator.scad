include <BOSL2/std.scad>;

$fn = 100;

slat_distance = 90;

slat_height = 22;
slat_thickness = 70;

module bed_slat_seperator(){
    
	cuboid([slat_distance,slat_thickness, slat_height], anchor=CENTER, rounding=-4, edges="Y");
}

bed_slat_seperator();
