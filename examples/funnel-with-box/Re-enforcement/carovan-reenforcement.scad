include <BOSL2/std.scad>


length = 25;
height = 30;
rounding= 2;

wedgeBodySize = [length, height, 30];

// Array holding configuration for each cylinder
cylinder_configs = [
    [20, 5, [0, -5, -4]],  // [height, radius, position]
    [30, 5, [20, 0, 0]]  // [height, radius, position]
];

cylinder_rotate = [120,0,0];


wedgeFaceAngle = [75,0,0];
wedgeFaceTranslate = [0,0,30]; 





    difference(){
        minkowski() {
            wedge(wedgeBodySize, center=true);
            sphere(r=rounding);
        }
        rotate(wedgeFaceAngle)
        translate(wedgeFaceTranslate)
        #wedge(wedgeBodySize+[40,30,20], center=true);
        
        // Add cylinders for screws
        for (config = cylinder_configs) {
            rotate(cylinder_rotate)
            translate(config[2])
            union(){
                cylinder(h=config[0], r=config[1], center=true);
                cylinder(h=config[0]+1000, r=config[1]-2, center=true);
            }
        }
    }
    