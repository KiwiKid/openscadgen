include <BOSL2/std.scad>
include <BOSL2/joiners.scad>


extra_clip_cutout_size =[21,7,20];
extra_clip_cutout_up = 12;
extra_clip_x_1 = 38;
extra_clip_x_2 = 92;
extra_clip_y = 21;

difference(){
    import("./trailer-plug-holder-with-cable-link.stl");

        back(extra_clip_x_1)
        up(extra_clip_cutout_up)
        right(-extra_clip_y)
        cuboid(extra_clip_cutout_size, anchor=CENTER);

        back(extra_clip_x_2)
        up(extra_clip_cutout_up)
        right(-extra_clip_y)
        cuboid(extra_clip_cutout_size, anchor=CENTER);
}