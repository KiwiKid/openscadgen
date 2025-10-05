
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;


designFileName = "football_cards";
initals = "GC";
name = "default";
version = "v0.1";
type = "insert";


module football_cards(initals="", type="insert"){


module initals(inital=initals, textDepth=textDepth){
        left((height/2)*0.80)
        fwd((width/2)*0.60)
        rotate([0,0,270])
		text3d(initals,h=textDepth,size=12, font = "Helvetica:style=Bold", center=true, anchor=BOTTOM);
}
	height = 100;
	width = 75;
    depth=0.7;
   // textDepth=0.3;

    if(type == "insert"){
        textDepth=0.4;
    
        difference(){
            cuboid([height,width,depth], rounding=5, edges="Z", anchor=BOTTOM);
            up(depth+0.001)
            down(textDepth)
            #initals(inital=initals, textDepth=textDepth);
        }
    }else {
    
        //textDepth=0.3;
            cuboid([height,width,depth], rounding=5, edges="Z");
            up(textDepth)
            initals(inital=initals, textDepth=textDepth);
		}
	//}
}

football_cards(initals=initals);
