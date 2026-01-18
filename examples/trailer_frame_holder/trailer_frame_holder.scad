

	include <BOSL2/std.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

    // Convenience helpers (local to this design).
    // mirrorX: mirror across the X-axis in "top view" (i.e. flip Y in 3D).
    module mirrorX() {
        mirror([0,1,0]) children();
    }

    // differenceX: subtract all following children from the first child,
    // and also subtract their mirrored copies via mirrorX().
    module differenceX() {
        difference() {
            children(0);
            for (i = [1 : $children-1]) children(i);
            mirrorX() for (i = [1 : $children-1]) children(i);
        }
    }

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    partType = "onf";
    
    blockers = "noBlockers";//includeBlockers



    strapThickness = 18;
    
    tralierBarWidth = 22.5;
    tralierBarHeight = 102; // 105
    tralierBarRounding = 5;
    
    holderWidth = 80;
    
    cutoutWidth = 140;
    cutoutHeight = 140;
    
    innerScrewRadius = 5;
    outerScrewRadius = 10;
    outerScrewDepth = -90;
    
     screwLineCount = 1;//3;
     
     
    screwLineSpacing = 90;
        screwOffset = 15;
        
    screwOut = 59;
    
    holderDepth = 30;
    
    
    module screw_holder(){
    right(tralierBarHeight)
    union(){
    rotate([0,90,0])
        cyl(r=innerScrewRadius, 4000);
        
   // rotate([0,90,0])
   //     outerScrewDepth)
    //    cyl(r=outerScrewRadius, tralierBarHeight*3, chamfer1=5);
     
        rotate([0,90,0])
        up(outerScrewDepth)
        linear_extrude(tralierBarHeight*3)
        hexagon(outerScrewRadius);
        }
    }
    
    screwCenterOffset = tralierBarHeight-3; //3
    
    screwBottomOffset = 35;
    
    
    

    // A line of screw holes on ONE side (positive Y). `differenceX()` will mirror it.
   
    module screw_hole_line() {
        for (i = [0 : screwLineCount-1]) {
            move([screwOut, screwOffset + i*screwLineSpacing, -screwOut])
                screw_holder();
        }
    }

	module trailer_frame_holder(){
        //cuboid([strapThickness, 2,2]);
        
        differenceX(){
            
            if(partType == "end"){
                left(11)
                cuboid([20,120,40], rounding=2);
            } else {
            cuboid([strapThickness+tralierBarHeight, tralierBarWidth+strapThickness+holderWidth,holderDepth], anchor=LEFT, rounding=30, edges=[RIGHT-BACK, RIGHT-FWD]);
            

            
            // Screw holes on both sides (mirrored) as a line.
           /* rotate([90,0,0])
            down(screwCenterOffset)
            #screw_hole_line();*/
            
            }
            
            left(0.01)
            back(10)
            cuboid([tralierBarHeight, tralierBarWidth,400], anchor=LEFT, rounding=tralierBarRounding, edges=RIGHT-FWD);
            
            
            
            rotate([90,0,0])
            fwd(holderDepth/2)
            up(screwCenterOffset)
            left(screwBottomOffset)
           back(0)
            screw_hole_line();
           
            
        } // differenceX

	}
blockerOffset = 45;
blockerHeight = 150;
blockerBaseHeight = 5;

    blockerSize = [blockerHeight,42,32];

    sliced(renderType=renderType) {
        if(blockers == "includeBlockers"){
        difference(){
        trailer_frame_holder();
        
        right(blockerHeight/2+blockerBaseHeight)
        fwd(blockerOffset)
        cuboid(blockerSize);
        
        right(blockerHeight/2+blockerBaseHeight)
        fwd(-blockerOffset)
        cuboid(blockerSize);
        
        }
        } else {
        
        trailer_frame_holder();
        }
    }
       








	
     
module sliced(
    renderType = "horzSlice",        // "horzSlice", "vertSlice", or "all"
    sliceSize = 1000,
    sliceThickness = 0.2,
    showRawSlices = false,
    horzSlicePos = [-500, -500, 0],
    vertSlicePos = [0, -500, -500]
) {
   
    module horz_slice(raw=false) {
        if (raw) {
            translate(horzSlicePos)
                cube([sliceSize, sliceSize, sliceThickness], center=false);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cube([sliceSize, sliceSize, sliceThickness], center=false);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cube([sliceThickness, sliceSize, sliceSize], center=false);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cube([sliceThickness, sliceSize, sliceSize], center=false);
            }
        }
    }

    if (renderType == "horzSlice") {
        horz_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "vertSlice") {
        vert_slice(raw=showRawSlices){
            children();
        }
    } else if (renderType == "all") {
        // show raw slices for reference
        horz_slice(raw=true);
        vert_slice(raw=true);
        // show full object
        children();
    } else {
        // show full object
        children();
    }
}

