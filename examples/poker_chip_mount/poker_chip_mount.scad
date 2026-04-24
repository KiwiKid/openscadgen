
include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

chipRadius = 19.2;
useSizer = false;
holderHeight = 1;

module poker_chip_mount(){
    difference(){
    if(useSizer){
        cyl(r=chipRadius+2 ,h=holderHeight);
    } else {
        cuboid([140,60,holderHeight], chamfer=0.5);
    }
     cyl(r=chipRadius ,h=10);
    }
}


poker_chip_mount();
