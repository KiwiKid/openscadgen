
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


initals = "GC";


module football_cards(initals=""){
	height = 100;
	width = 75;
	difference(){
		cuboid([height,width,0.1], rounding=5, edges="Z");
		left((height/2)*0.85)
        fwd((width/2)*0.7)
        rotate([0,0,270])
		text3d(initals,h=height,size=7, font = "Baskerville:style=Bold", center=true);
		
	}
}

football_cards(initals=initals);
