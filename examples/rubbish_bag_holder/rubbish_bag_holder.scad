

	include <BOSL2/std.scad>;

include <BOSL2/walls.scad>;
include <BOSL2/screws.scad>;
	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 200;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    

    wallSize = 3;

    rubbishBagRadius = 50;
    rubbishBagLength = 150;
    
    
    flatSectionWidth = rubbishBagRadius*2.3;
    flatSectionLength = rubbishBagRadius*2;
    
    backWallThickness = 3;
    
    slotLengthReduce = 9.95;
    slotSize = [10, 4, rubbishBagLength-slotLengthReduce];
    slotSize2 = [10*1.1 , 4*1.1, rubbishBagLength-slotLengthReduce];
   
   supportRotate = 0;
   supportSize = [5,1,rubbishBagLength*0.9];
   supportToCenter = 2;
   
   topSupportHeight = 10;
    
    
    slotEndWallSize = 5;
    slotMove = [0,rubbishBagRadius-2,slotEndWallSize];
    slotRotate = 40;
    
    flatOffset = 6;
    
    screwOffset = 25;
    
    patternRepeatCount = 8;
    patternRotate = 18;
    patternCenterOffset  = rubbishBagRadius*2.5;//rubbishBagRadius*2.165;
    intersectBorderBuffer = 115;
    
    topSupportWidth = 100;
    
    
    edgeDepth = 15;
    wedgePointness  = 5;
    
    teethMove = [-wedgePointness,0,0];
    teeth2Move = [wedgePointness,0,-3];
    
    
    
    module screwHole(){
    
    screwdriverHoleHeight = rubbishBagRadius*2.3;
    screwdriverHoleRadius = 4;
    
        yrot(90){
        up(2.5)
        cyl(r=1.8, h=3.5, chamfer1=-1);
        
        // screwdriverhole
        down(screwdriverHoleHeight/2)
        cyl(r=screwdriverHoleRadius, h=screwdriverHoleHeight);
        }
    
    }
    
    module ring(radius=10, height=10, ringWidth=3){
    left(radius)
        difference(){

            
            cyl(r=radius, h=height);
            cyl(r=radius-ringWidth, h=height);
        }
    }
    
    module rod(left_handed=false){
     threaded_rod(
        d = rubbishBagRadius*2.1,      // 100mm Outer Diameter
        l = rubbishBagLength,       // Length
        pitch = 10,    // Size of individual ridges (easy to print)
        starts = 4,    // 4 parallel tracks! Effective pitch (lead) is 24mm!
        left_handed = left_handed
        );
    }
    
    module teeth(){
    teethCount = 20;
    teethHeight = 20;
    teethBuffer = 13;
    xrot(90)
    yrot(90)
        ycopies((rubbishBagLength-teethBuffer)/teethCount, n=teethCount){
        xrot(0)
        wedge([2, 5, wedgePointness]);
        }
    
    }

	module rubbish_bag_holder(){
        intersection(){
        difference(){
                cyl(r=rubbishBagRadius, h=rubbishBagLength);
        		cyl(r=rubbishBagRadius-wallSize, h=rubbishBagLength+1);
                
                zrot(slotRotate)
                move(slotMove)
               cuboid(slotSize);
                
                
                
                
              }
        left(rubbishBagRadius/flatOffset)
        cuboid([flatSectionWidth, flatSectionLength, rubbishBagLength]);
        
        union(){
        
       zrot(45)
     rod();
     
     zrot(45)
     rotate([-180,0,0])
     rod(left_handed=true);
    
    up(rubbishBagLength/2)
    cyl(r=rubbishBagRadius, h=edgeDepth);
    
    
    down(rubbishBagLength/2)
    cyl(r=rubbishBagRadius, h=edgeDepth);
    }
        
        /*move(pivot)
        rot_copies(n=patternRepeatCount) {
            right(patternCenterOffset) 
            intersection(){
            sparse_wall(h=rubbishBagLength, l=50*3, thick=rubbishBagLength, strut=4, maxang=45);
               cuboid([100*3-intersectBorderBuffer,rubbishBagLength-intersectBorderBuffer,rubbishBagLength]);
            }
        }*/

// Rotate a cube 45 degrees around that pivot point



// Mark the pivot point so you can see it
//move(pivot) sphere(d=3); 
        
        }
        
                difference(){
                
                zrot(slotRotate)
                move(slotMove)
               cuboid(slotSize2, chamfer=0.1);
               
               
                zrot(slotRotate)
                move(slotMove)
                up(2)
                scale([0.5,1.5,1])
                 cuboid(slotSize);
                 
                      }
               zrot(slotRotate)
                move(slotMove+teethMove)
                 teeth();
                 
                 zrot(slotRotate)
                move(slotMove+teeth2Move)
                zrot(180)
                 teeth();
   
        
           /*     zrot(slotRotate)
                move(slotMove)
                fwd(supportToCenter)
                zrot(-supportRotate)
               cuboid(supportSize, rounding=0.5, edges=[FWD,BACK]);
               */
        
         intersection(){
            cyl(r=rubbishBagRadius, h=rubbishBagLength);
            
            right(rubbishBagRadius-2.2)
            
            
            
            up(rubbishBagLength/2-topSupportHeight/2)
            ring(radius=rubbishBagRadius, height=topSupportHeight, ringWidth=topSupportWidth)
            
            down(rubbishBagLength/2-topSupportHeight/2)
            cyl(r=rubbishBagRadius, height=topSupportHeight, ringWidth=topSupportWidth);
         }
         
         
            right(rubbishBagRadius-2.4)
         difference(){
            cuboid([backWallThickness,20,rubbishBagLength]);
            up(screwOffset)
                screwHole();
                
                down(screwOffset)
                screwHole();
            }
	}


    sliced(renderType=renderType) {
        rubbish_bag_holder();
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
                cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(horzSlicePos)
                    cuboid([sliceSize, sliceSize, sliceThickness], anchor=[-1,-1,-1]);
            }
        }
    }

    module vert_slice(raw=false) {
        if (raw) {
            translate(vertSlicePos)
                cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
        } else {
            intersection() {
                children();
                translate(vertSlicePos)
                    cuboid([sliceThickness, sliceSize, sliceSize], anchor=[-1,-1,-1]);
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

