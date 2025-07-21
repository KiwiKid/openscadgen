include <BOSL2/std.scad>
$fn = 100;

// Parameters
length = 50;
width = 20;
thickness = 2;
tab_length = 5;
tab_width = 8;
screw_dia = 3;
bulge_scale = 1.5;
bulge_height = thickness;
screw_depth = 7;
// Full model
module batteryClip() {
  // Base cover plate
  cuboid([length, width, thickness], anchor=BACK+LEFT);


  difference(){
      scale([bulge_scale, 1, 1])
        move([0,-width/2,0])
        cyl(h=bulge_height, d=width);

           move([-screw_depth,-width/2,0])
      cyl(h=thickness + 2, d=screw_dia);
}
    // Tab on LEFT
    move([length,-width/2,0])
      cuboid([tab_length, tab_width, thickness/2], anchor=LEFT+TOP);
  }

batteryClip();