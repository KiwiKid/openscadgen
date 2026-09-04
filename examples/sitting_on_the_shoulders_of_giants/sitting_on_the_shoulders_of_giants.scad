include <BOSL2/std.scad>;

$fn = 100;

// Use "horzSlice" or "vertSlice" to inspect a printable cross-section,
// "all" to show the object and both slice planes, or "obj" for the model.
renderType = "obj";
globalScale = [0.5,0.5,0.5];

// The STL files use different local origins. These values place the seated
// figure on the Hulk's left shoulder while keeping the shoulder as a compact,
// printable display plinth. Tweak these in the Customizer for a new pose.
/* [Figure placement] */
manRight = 150;
manOut = 410;
manUp = 200;
man_position = [manRight, manOut, manUp];
man_rotation = [0, 0, -90];
man_scale = 1;
man_size = [40, 40, 73];

/* [Hulk shoulder cut] */
hulk_size = [204, 126, 280];
shoulder_box_center = [320, 320, 300];

shoulder_box_height = 250;

shoulder_box_width = 200;
shoulder_box_size = [shoulder_box_width, 400, shoulder_box_height];


seatPosition =  [100,110,17];
seatSize = [20,30,20];

module seated_man() {
    translate(man_position)
        rotate(man_rotation)
            scale(man_scale)
                resize(man_size)
                    difference(){
                    import("siting-man.stl", convexity=10);
                    // cutout seat
                    }
}

module hulk_shoulder() {
    // Cut the shoulder region from the resized bust with a configurable box.
    intersection() {
        resize(hulk_size)
            import("hulk-bust.stl", convexity=10);
        translate(shoulder_box_center)
            cuboid(shoulder_box_size, anchor=CENTER);
    }
}

module sitting_on_the_shoulders_of_giants(mode="onlyMan") {
if(mode == "full"){
    union() {
    
        hulk_shoulder();
        seated_man();
    }
} else if(mode == "onlyMan"){
        seated_man();   
        }
        }
        

include <../../openscad-lib/openscadgen-core.scad>;

sliced(
    renderType=renderType,
    horzSlicePos=[-500, -500, 80],
    vertSlicePos=[105, -500, -500]
) {
    scale(globalScale)
    sitting_on_the_shoulders_of_giants(mode="full");
}
