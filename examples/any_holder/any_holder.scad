include <BOSL2/std.scad>;

$fa = .01;
$fs = $preview ? 5 : 1;
$fn = 200;

thingFrameRounding = 2;
insetThing = "square"; // square round


/// Frame size
frameWallWidth = 5;
frameDepth = 5;

insetPrimarySize = 49;
insetSecondarySize = 49;

// FACEPLATE STYLE
frameWidth = insetPrimarySize + frameWallWidth;
frameHeight = insetSecondarySize + frameWallWidth;
squareFrameSize = [frameDepth,frameWidth,frameHeight];


insetDepth = 6;

insetOffset = frameDepth/2-insetDepth/2+0.01;

// square inset settings
squareInsetSize = [insetDepth, insetPrimarySize, insetSecondarySize];
insetRounding = 8;
smallHookMove = [10, -5,0];
largeHookMove = [15, -5, 0];

// round settings

roundHolderSize = 22;
roundInsertRadius = 20;
roundInsertSize = 20;
roundInsertDepth = 10;

/// ATTACHMENT STYLE
attachmentType = "largeHook"; // "smallHook" "base" largeHook
hookType = "central"; // "full" "right" "central"
hookRotate = 20;

// smallHook only settings
smallHookRadius = 12;
smallRoundHookMove = [2.5, 2, 0];
largeRoundHookMove = [10, 0, 0];

// largeHook only settings
largeHookRadius = 20;

hookUp = 2;



smallHookRoundMove = [-roundHolderSize+10,0,roundHolderSize-10];

hookReduceCube = [largeHookRadius*5,frameWidth/2,largeHookRadius*3];
reduceCubeOffset = 13;

lineSizeOneMove = [0,0,2];
lineSizeTwoMove = [0,0,0];

lineSize = [1,frameWidth/2,1];


hookStartRadius = 50;
include <BOSL2/std.scad>

// Example usage:
// This matches your criteria: Flat horizontal alignment, 40-degree exit angle, and 40-degree twist

// Example usage:
// This arches directly forward (along the Y-axis) and curves up into the air
include <BOSL2/std.scad>

// Example usages:
// 1. Tall, steep arch

// 2. Low, flat, stretched arch (uncomment to test)
// translate([0, 80, 0])
//     rainbow_bezier_hook(arch_height=20, arch_span=80, twist_deg=40);


module rainbow_bezier_hook(
    base_size    = [100,10], // Uniform [width, thickness] profile
    corner_rad   = 1.5,     // Smooth rounding radius for corners
    arch_span    = 40,      // Total distance the arch travels forward (Y-axis)
    arch_height  = 40,      // Peak height of the arch in the air (Z-axis)
    twist_deg    = 40,      // Continuous twist amount across the arch length
    steps        = 60,      // Path resolution (higher = smoother curve)
    fn_val       = 64       // Resolution for the profile shape
) {
    $fn = fn_val;

    // 1. Generate the 2D cross-section data path
    shape_profile = rect(base_size, rounding=corner_rad, center=true);

    // 2. Define Bézier Control Points in the YZ plane (X is locked at 0)
    // P0: Start on floor, P1: Upward tangent handle, P2: Forward tangent handle, P3: End on floor
    control_points = [[0,0],                         // Start point
        [0, 0, arch_height * 1.5],         // Shoot straight up to force vertical entry
        [0, arch_span, arch_height * 1.5], // Stay high over the end span
        [0, arch_span, 0]                  // End point landing back on floor level
    ];

    // 3. Generate the smooth adjusted curve from the control points
    spine_path = bezier_curve(control_points, splines=steps);

    // 4. Sweep the profile over the custom adjusted spine
    path_sweep(shape_profile, spine_path, twist=twist_deg, scale=1.0);
}






module hook(hookRadius=6, hookMove=hookMove, hookType="full"){
    intersection(){
    if(hookType == "full"){
    
            fwd(reduceCubeOffset)
        cuboid(hookReduceCube*10);
    } else if(hookType == "left"){
            back(reduceCubeOffset)
            cuboid(hookReduceCube);
   
    }else if(hookType == "right"){
          cuboid(hookReduceCube);
    } else if(hookType == "central"){
    
          cuboid(hookReduceCube);
       }
       
       half_of([3,0,20], 2) {
    rotate([90,0,0])
    
        move(hookMove)
        //zrot(30)
        difference(){
            cyl(r=hookRadius,h=frameWidth, rounding=4);
            
            fwd(1)
            ycopies(2, n=3){
               cyl(r=hookRadius-1.6,h=frameWidth+1);
               
            }
                right(1)
               cyl(r=hookRadius-1.6,h=frameWidth+1);
        }
        }
        
        
        }

        left(frameDepth/2)
        hookConnector();
}

module attachment(attachmentType="largeHook", smallHookMove=smallHookMove, largeHookMove=largeHookMove){
    if(attachmentType == "largeHook"){
        position(RIGHT+TOP) yrot(180) xrot(-90) up(hookUp) 
        //hook(hookRadius=largeHookRadius, hookMove=largeHookMove, hookType=hookType);
        left(hookStartRadius)
rainbow_bezier_hook(arch_height=60, arch_span=50, twist_deg=40);

        
    } else if(attachmentType == "smallHook"){
        position(RIGHT+TOP)  zrot(-hookRotate) up(hookUp) hook(hookRadius=smallHookRadius, hookMove=smallHookMove, hookType=hookType);
    } else {
        echo("!!!!!!!! attachmentType not found");
    }
}

module holder(insetThing="square"){
    if(insetThing == "square"){
        cuboid(squareFrameSize, rounding=thingFrameRounding) children();
    } else if(insetThing == "round"){
        rotate([0,90,0])
        cyl(r=roundHolderSize, h=frameDepth, rounding=1)
         rotate([0,-90,0])
         move(smallHookRoundMove)
         children();
    } else {
        echo("!!!!!!!!! insetThing is not valid");
    }
}

module any_holder(){
    difference(){
        // GROUP 1: The solid positive shapes to be joined together
        // Enclosing these ensures the holder + attachment pass through cleanly
        union() {
            holder(insetThing=insetThing)
                if(insetThing == "square"){
                attachment(attachmentType=attachmentType, largeHookMove=largeHookMove, smallHookMove=smallHookMove);
                } else {
                attachment(attachmentType=attachmentType, largeHookMove=largeRoundHookMove, smallHookMove=smallRoundHookMove);
               }
        }
        
        // GROUP 2: The cutters (everything below this point gets subtracted)
        if(insetThing == "square"){
            left(insetOffset)
                cuboid(squareInsetSize, rounding=insetRounding, edges="X");
                
                // cord hole
                down(frameWidth/2)
                right(2)
                cuboid([5,15,10]);
                
        } else if(insetThing == "round"){
        
           rotate([0,90,0])
           //down(roundInsertDepth)
                cyl(r=roundInsertRadius, h=insetDepth+3, rounding1=-2);
                
                  // cord hole
                //down(roundInsertSize)
                //right(-roundInsertDepth*1.3)
                //#cuboid([5,15,10]);
                
        } else {
            echo("!!!!!!!!! insetThing is not valid");
        }
    }
}

any_holder();