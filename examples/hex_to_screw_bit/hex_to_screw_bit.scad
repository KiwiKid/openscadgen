include <BOSL2/std.scad>;

$fn = 100;

screwRadius = 1.16;
screwDepth = 15;
hexLength = 30;
hexRadius = 7.2/2;

setScrewLength = 26;
setScrewHeight = 5;
setScrewRadius = 0.8;

module hex_to_screw_bit(){
	difference(){
        linear_extrude(hexLength)
        hexagon(r=hexRadius, realign=true);
        
        cyl(r=screwRadius, h=screwDepth, anchor=BOT, chamfer1=-.5);
        
        up(setScrewHeight)
        xcyl(r=setScrewRadius, h=setScrewLength, anchor=LEFT);
    }
}

hex_to_screw_bit();
