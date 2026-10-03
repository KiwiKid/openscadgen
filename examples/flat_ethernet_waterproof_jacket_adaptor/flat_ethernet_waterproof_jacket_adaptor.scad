include <BOSL2/std.scad>;

$fn = 100;

height = 30;
radius = 6;

module flat_ethernet_waterproof_jacket_adaptor(){
	difference(){
        cyl(r=radius, h=height);
        fwd(1)
        cuboid([2,radius*2,height+1]);
    
    }
}

flat_ethernet_waterproof_jacket_adaptor();
