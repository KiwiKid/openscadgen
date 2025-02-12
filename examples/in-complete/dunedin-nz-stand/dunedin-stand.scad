$fn = 2;
 include <BOSL2/std.scad>


part = "both";

mapRotate = [0, -70, 0];


frameHeight = 82;
frameWidth = 82;
frameBorder =8;
frameDepth = 3;

if (part == "map" || part == "both"){
    translate([-55,-140,-83])
    rotate(mapRotate)
    difference(){
        import("./otago_peninsula_full.stl");
        translate([140, 140, 0])
    rect_tube(size=[frameWidth+2,frameHeight+2], wall=frameBorder, rounding=10, ichamfer=7, h=10);
    }
}

 if(part == "stand" || part == "both") {
    // Create a cone using the cylinder function
    // A cone is a cylinder with a top radius of 0
    translate([0, 0, 4.8])
    difference() {
        translate([0, 0, 22.8])
        cylinder(h = 55, r1 = 20, r2 = 0, center = true);
        
        // Subtract the text from the bottom
        translate([-100, -100, -30])  // Adjust to fit the cylinder
        rotate([180, 180, 180])
        linear_extrude(height=2)  // Set depth of text engraving
        text("Otago Peninsula", size=3);
    }
   
   translate([-5,-frameHeight/2+41,41])
    rotate(mapRotate) 
    cube([frameWidth-frameBorder-3,frameHeight-frameBorder-3, frameDepth], true);
    
    translate([-5, 0, 41])
    rotate(mapRotate)
    rect_tube(size=[frameWidth+2,frameHeight+2], wall=frameBorder, rounding=10, ichamfer=7, h=frameDepth);
}
// You can adjust the height (h), base radius (r1), and position as needed
