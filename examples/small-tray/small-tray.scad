include <BOSL2/std.scad>

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

tray_width = 100;
tray_height = 200;
tray_step = 2;
wall = 2;
height = 25;
pipe_radius = 10;
pipe_width = 2;
part_id_letter = "D";

connection_length = 6;

module tray(){

difference() {
    cuboid([tray_width, tray_height, height], rounding=10, edges=[BOTTOM+LEFT, BOTTOM+RIGHT, LEFT+FWD, BOTTOM+FWD, BOTTOM+BACK,RIGHT+FWD, RIGHT+BACK, LEFT+BACK]);
    
    up(wall)
    cuboid([tray_width-wall*2, tray_height-wall*2, height], rounding=10, edges=[BOTTOM+LEFT, BOTTOM+RIGHT, LEFT+FWD, BOTTOM+FWD, BOTTOM+BACK, RIGHT+FWD, RIGHT+BACK, LEFT+BACK]);
    
    //#back_half()
    fwd(tray_height/2-30)
    up(height/2-tray_step)
    rotate([90,0,0])
    tube(ir=0,or=pipe_radius, h=100);
    
    fwd(-tray_height/2)
    up(height/2)
    rotate([90,0,0])
    front_half()
    tube(ir=0,or=pipe_radius, h=connection_length);
    
        
}

    fwd(-tray_height/2)
    up(height/2)
    rotate([90,0,0])
    front_half()
    tube(ir=pipe_radius-pipe_width,or=pipe_radius, h=connection_length);
    }

    
font_size = 10;
engrave_depth = 10;



difference() {
    tray();
    translate([tray_width/2, 0, tray_height/2 - engrave_depth+10])
    rotate([90,0,90])
    #text3d(part_id_letter, h=3, size=10);
}

