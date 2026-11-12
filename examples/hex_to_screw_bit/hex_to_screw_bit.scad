include <BOSL2/std.scad>;

$fn = 100;

screwRadius = 0.8;
screwDepth = 8;
hexLength = 20;
hexRadius = 6.5/2;

setScrewLength = 18;
setScrewHeight = 5;
setScrewRadius = 0.7;

module hex_to_screw_bit(){
	difference(){
        linear_extrude(hexLength)
        hexagon(r=hexRadius, realign=true);
        
        cyl(r=screwRadius, h=screwDepth, anchor=BOT);
        
        up(setScrewHeight)
        xcyl(r=setScrewRadius, h=setScrewLength, anchor=LEFT);
    }
}

hex_to_screw_bit();
