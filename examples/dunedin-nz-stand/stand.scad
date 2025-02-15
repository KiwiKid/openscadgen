$fn = 200;
 include <BOSL2/std.scad>

mapRotate = [0, -70, 0];

mapSTLScale = 0.96;
mapSTLScaling = [mapSTLScale, mapSTLScale, mapSTLScale];

frameRimHeight = 2;

// SHARED WITH MAP.scad
frameHeight = 84;
frameWidth = 84;
frameBorder =8;
frameDepth = 3;
frameRounding = 10;
frameIchamfer= 7;



    translate([0, 0, 4.8])
    difference() {
        translate([0, 0, 22.8])
        cylinder(h = 55, r1 = 20, r2 = 0, center = true);
        
        // Subtract the text from the bottom 
        translate([-100, -100, -30])  
        rotate([180, 180, 180])
        linear_extrude(height=2)  
        text("Otago Peninsula", size=3);
    }
 
 // Back plate
translate([-5,-frameHeight/2+41,41])
rotate(mapRotate) 
cube([frameWidth-frameBorder-3,frameHeight-frameBorder-3, frameDepth/2], true);

translate([-5, 0, 41])
rotate(mapRotate)
rect_tube(size=[frameWidth, frameHeight], wall=frameBorder, rounding=frameRounding, ichamfer=frameIchamfer, h=frameDepth);

