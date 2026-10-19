include <BOSL2/std.scad>;

$fn = 100;

holeStartDepth = 4;
bitHeight = 15;

bit_size = 6.55; // 6.35mm nominal + tolerance

bitRadius = 2.5;

holeRadius = 0.6;

screwHolderRadius = .5;

screwHolderOffset = 4;

module hex_drill_bit_holder(){
difference(){
    intersection(){
    linear_extrude(h=bitHeight)
	hexagon(ir=bit_size/2);
    cuboid([bit_size,bit_size,bitHeight], anchor=BOT, chamfer=3, edges=TOP);
    }
    
    up(holeStartDepth){
    cyl(r=holeRadius, h=1000, anchor=BOT);
    
    up(screwHolderOffset)
       ycyl(r=screwHolderRadius, h = 100, anchor=FWD);
    }
    
    
    }
}

//left_half()
hex_drill_bit_holder();
