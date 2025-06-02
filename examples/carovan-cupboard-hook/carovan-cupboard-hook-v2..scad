include <BOSL2/std.scad>

$fn =20;

module carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move, difference_size,
    holder_size, holder_width, hook_move, 
    holder_cutout_move, holder_cutout_size, 
    holder_cutout_2_move,
    hook_angle, hook_count,
     hook_gap
) {
    module base_shape() {
    
        module hook(xOffset = 0){
            left(xOffset)
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
        }
        
        if(hook_count > 0){
            for ( i = [0 : hook_count] ){
               hook((i*hook_gap));
            }
        }


        // Door hook
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

 height = !is_undef(height) ? height : 9 ;
 rounded = !is_undef(rounded) ?  rounded : "true";
 hook_angle = !is_undef(hook_angle) ?  hook_angle : 50;

connector_move = [8, 30, 0];
connector_size = [150, 24];

holder_width=90;

hook_depth = 85;
hook_y_offset = 75;
hook_count =6;


// Debug view
carovan_hook_shape(
    r1, r2, R, length, height, rounded,
    connector_move, connector_size, difference_move=[0, 40, 0], difference_size=[-2, -5],
    holder_size=60, holder_width=holder_width, hook_move= [53, hook_y_offset, 40], 
    holder_cutout_move= [-40, 5, 0], holder_cutout_size= [holder_width+30, holder_width-20], 
    holder_cutout_2_move= [-hook_depth, 10, 0],
    hook_angle=hook_angle, hook_count=hook_count,
     hook_gap=connector_size[0]
);
