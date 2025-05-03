include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

tray_width = 230;
tray_height = 250;
tray_step = 0;
wall = 1;
height = 28;
pipe_radius = 5;
pipe_inner_radius = 2;
pipe_width = 5;
pipe_wall = 1;
part_id_letter = "D";

connection_length = 4;
connection_offset = 0.6;
pipe_height = 4;



partType = "box"; //startBox, endBox, box

module tray(){

difference() {
    cuboid([tray_width, tray_height, height], rounding=10, edges=[BOTTOM+LEFT, BOTTOM+RIGHT, LEFT+FWD, BOTTOM+FWD, BOTTOM+BACK,RIGHT+FWD, RIGHT+BACK, LEFT+BACK]);
    
    up(wall)
    cuboid([tray_width-wall*2, tray_height-wall*2, height], rounding=10, edges=[BOTTOM+LEFT, BOTTOM+RIGHT, LEFT+FWD, BOTTOM+FWD, BOTTOM+BACK, RIGHT+FWD, RIGHT+BACK, LEFT+BACK]);
    
    if( partType ==  "startBox" || partType == "box"){
    // holder conection
        fwd(tray_height/2+pipe_height/2-wall-0.001)
        up(height/2)
        rotate([90,0,0])
        tube(ir=0,or=pipe_radius, h=pipe_height);
    }
        if( partType ==  "endBox" || partType == "box"){
        // Rimmed Connection
            fwd(-tray_height/2+0.002)
            up(height/2)
            rotate([90,0,0])
            front_half()
            tube(ir=0,or=pipe_radius, h=connection_length);
            
            
        }
    
        
}
 if( partType ==  "endBox" || partType == "box"){
 difference(){
    fwd(-tray_height/2)
    up(height/2)
    rotate([90,0,0])
    front_half()
    #tube(ir=pipe_inner_radius,or=pipe_radius, h=connection_length);
    
    fwd(-tray_height/2+0.002)
        back(connection_offset)
        up(height/2)
        rotate([90,0,0])
        front_half()
        tube(ir=pipe_inner_radius+(wall*2),or=pipe_radius+0.001, h=pipe_wall);
        }
    
    
    }
    }

    
font_size = 10;
engrave_depth = 10;



difference() {
    tray();
    translate([tray_width/2, 0, tray_height/2 - engrave_depth+10])
    rotate([90,0,90])
    #text3d(part_id_letter, h=3, size=10);
}

