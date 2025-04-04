include <BOSL2/std.scad>

$fn =20;

module carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move, difference_size,
    holder_size, holder_width, hook_move, 
    holder_cutout_move, holder_cutout_size, 
    holder_cutout_2_move,
    hook_angle
) {
    module base_shape() {
        difference() {
            union() {
                rotate([0, 0, -hook_angle])
                difference() {
                    egg(length, r1, r2, R, $fn = 180);
                    move([5, 10, 0])
                    egg(length, r1, r2, R, $fn = 180);
                }
                move(connector_move)
                rect(connector_size, 2);
            }

            move(difference_move)
            rect(connector_size - difference_size);
        }

        
        // Hook
        move(hook_move)
        difference() {
            rect([holder_size, holder_width], 2);
            move(holder_cutout_move)
            rect(holder_cutout_size);

            move(holder_cutout_2_move)
            rect(holder_cutout_size);
        }
    }

    if (rounded == "true") {
        corner_radius = 2;
        minkowski() {
            scale(.7)
            linear_extrude(height = height)
            base_shape();

            sphere(r = corner_radius, $fn = $fn);
        }
    } else {
        base_shape();
    }
}



// Define parameters
r1 = 25; 
r2 = 12; 
R = 65;
length = 70;

 height = !is_undef(height) ? height : 10 ;
 rounded = !is_undef(rounded) ?  rounded : "true";
 hook_angle = !is_undef(hook_angle) ?  hook_angle : 50;

connector_move = [8, 30, 0];
connector_size = [60, 24];

hook_depth = 55;


// Debug view
carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move=[0, 40, 0], difference_size=[-2, -5],
    holder_size=40, holder_width=35, hook_move= [52, 55, 0], 
    holder_cutout_move= [-40, 0, 0], holder_cutout_size= [110, 27], 
    holder_cutout_2_move= [-hook_depth, 10, 0],
    hook_angle=hook_angle
);
