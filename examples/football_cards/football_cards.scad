
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


designFileName = "football_cards";
initals = "GC";
name = "default";
version = "v0.1";


module football_cards(initals=""){


module initals(inital=initals){
        left((height/2)*0.85)
        fwd((width/2)*0.7)
        rotate([0,0,270])
		text3d(initals,h=depth,size=7, font = "Helvetica:style=Bold", center=true);
}
	height = 100;
	width = 75;
    depth=0.5;

        difference(){
            union(){
            cuboid([height,width,depth], rounding=5, edges="Z");
            up(0.2)
            initals(initals);
        }
         down(0.2)
                initals(initals);
        }
		
	//}
}

football_cards(initals=initals);
