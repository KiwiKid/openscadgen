include <BOSL2/std.scad>

$fn =20;

module carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move, difference_size,
    holder_size, holder_width, holder_move, 
    holder_cutout_move, holder_cutout_size, 
    holder_cutout_2_move
) {
    module base_shape() {
        difference() {
            union() {
                rotate([0, 0, -40])
                difference() {
                    egg(length, r1, r2, R, $fn = 180);
                    move([2, 11, 0])
                    egg(length, r1, r2, R, $fn = 180);
                }
                move(connector_move)
                rect(connector_size, 2);
            }

            move(difference_move)
            rect(connector_size - difference_size);
        }

        
        // Hook
        move(holder_move)
        #difference() {
            rect([holder_size, holder_width], 2);
            move(holder_cutout_move)
            rect(holder_cutout_size);

            left(10)
            back(20)
            move(holder_cutout_move)
            rect(holder_cutout_size);
        }
    }

    if (rounded) {
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
height = 80;
rounded = true;

connector_move = [15, 28.4, 0];
connector_size = [65, 20];
difference_move = [0, 36, 0];
difference_size = [-2, -5];

holder_cutout_2_move = [10, 10, 0];

// Render the shape
/*carovan_hook_shape(
    r1, r2, R, length, height, false,
    connector_move, connector_size, difference_move, difference_size,
    holder_size=35, holder_width=30, holder_move= [60, 38, 0], 
    holder_cutout_move=[-23, 0, 0], holder_cutout_size= [70, 23], 
    holder_cutout_2_move
);
*/
// Define parameters
r1 = 25; 
r2 = 12; 
R = 65;
length = 80;
height = 10;
rounded = true;

connector_move = [20, 18.4, 0];
connector_size = [90, 20];

holder_width = 30;
holder_cutout_2_move = [10, 10, 0];

// Debug view
carovan_hook_shape(
    r1, r2, R, length, height, true,
    connector_move, connector_size, difference_move=[10, 33, 0], difference_size=[-2, -5],
    holder_size=70, holder_width=60, holder_move= [100, 45, 0], 
    holder_cutout_move= [-50, 0, 0], holder_cutout_size= [150, 40], 
    holder_cutout_2_move
);
