include <BOSL2/std.scad>

$fn =20;

module carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move, difference_size,
    holder_size, holder_width, holder_move, 
    holder_cutout_move, holder_cutout_size, 
    holder_cutout_2_move, hook_angle, hook_move, hook_connector_height, hook_connector_width, hook_connector_size
) {
    module base_shape() {
        difference() {
            union() {
                   move(hook_move)
                rotate([0, 0, -hook_angle])
                difference() {
                    egg(length, r1, r2, R, $fn = 180);
                    move([5, 12, 0])
                    egg(length, r1, r2, R, $fn = 180);
                }
                move(connector_move)
                rect(connector_size, 2);
            }

            move(difference_move)
            #rect(connector_size - difference_size);
        }

        
        // Hook
        move(holder_move)
        difference() {
        union(){
            rect([holder_size, holder_width], 2);
            fwd(10)
            left(hook_connector_height)
            rect([hook_connector_size, hook_connector_width], 2);
           }
            move(holder_cutout_move-[holder_size, 0, 0])
            rect(holder_cutout_size);

            // Hook difference block
            left(30)
            back(10)
            move(holder_cutout_move_2)
            rect(holder_cutout_size+[0,3,0]);
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
height = 1;
rounded = true;

connector_move = [20, 25.4, 0];
connector_size = [60, 9];

holder_width = 25;
holder_cutout_move_2 = [-5, 0, 0];

hook_angle = 30;
hook_move=  [18,5,0];

// Debug view
carovan_hook_shape(
    r1, r2, R, length, height, true,
    connector_move, connector_size, difference_move=[10, 37.2, 0], difference_size=[-2, -5],
    holder_size=33, holder_width=35, holder_move= [60, 60, 0], 
    holder_cutout_move= [13, 0, 0], holder_cutout_move_2, holder_cutout_size= [70, 28], 
    holder_cutout_2_move, hook_angle, hook_move, hook_connector_height=13.7, hook_connector_width = 55, hook_connector_size = 7
);
