

	include <BOSL2/std.scad>;
include <BOSL2/joiners.scad>;

	$fa = .01;
	$fs = $preview ? 5 : 1;
	$fn = 20;

	/*
	renderType:
	use to print a test slice and confirm sizing before printing:
	 - "horzSlice" - horizontal slices (default)
	 - "vertSlice" - vertical slices
	 - "all" - the whole object
	*/
	renderType = "obj";
    
    partType = "box"; // "frontWall" "sideWall" "box"
    
    boxHeight = 60;
    boxWidth = 38;
    boxDepth = 63;
    wallSize = 2;
    slideGrooveDepth = 2;
    
        sideWallSize = [wallSize, boxWidth-(wallSize*2)+slideGrooveDepth, boxDepth];
                sideWallSlideSideOffset = 3;

sideWallMove = [(boxHeight/2)-sideWallSlideSideOffset, 0, 0];

frontWallSlideOffset = 3;

grooveSlackMultipler = 1.01;

        frontWallSize = [boxHeight-sideWallSlideSideOffset-wallSize*2, boxWidth-wallSize*2+slideGrooveDepth, wallSize];
      frontWallGrooveSize = [boxHeight, boxWidth-wallSize, wallSize*grooveSlackMultipler];
      
        frontWallMove = [0,0,boxDepth/2-frontWallSlideOffset];
        
        
        buttonBoxSize = [boxHeight, boxWidth, boxDepth];
        buttonBoxInsertSize = [boxHeight, boxWidth*0.9, boxDepth*1.4];
        
        
        
        //buttonHoleSize = [50.1,30.5,10.0];
        buttonHoleRadius = 12.3;
        
        dovetailFwdBlockSize = [boxHeight, boxDepth, 10];
        
        
        holeRadius = 10;
     module box_cord_hole(){
        xrot(-35)
        ycopies(4,8){
            cyl(r=holeRadius, h=300);
        }
        up(boxHeight/2)
        cuboid([3,boxHeight, 130]);
     
     }

    module  front_wall(frontWallSize=frontWallSize, includeHole=false){
        if(includeHole){
            difference(){
            up(boxDepth/2)
            cuboid(frontWallSize, anchor=BOT);
            
            up(boxDepth/2)
            down(5)
            #cyl(r=buttonHoleRadius, h=10, anchor=BOT);
            //cuboid(buttonHoleSize, anchor=BOT);
            
            }
        }else {
            up(boxDepth/2)
            cuboid(frontWallSize);
        }
        }
    
    module dovetail_core(type="side", sex="male"){
        if(type == "side"){
            
            dovetail(sex, slide=boxDepth, h=5, w=30);
        }
        }
        
    module box_core(){
        cuboid(buttonBoxSize, anchor=BOTTOM) {
        attach(LEFT)  dovetail_core();
            

            attach(BACK) dovetail_core();
            attach(FWD) 
            difference(){
                cuboid(dovetailFwdBlockSize);
                up(5)
                dovetail_core(sex="female");
                
                }
               }
       
    
    }
    
    module box(){
        difference(){
        box_core();
        //    cuboid(buttonBoxSize, anchor=BOTTOM); 
            rotate([90,0,90])
            move([0, 0, -boxHeight/2-20])
            box_cord_hole();
            
            move([wallSize, 0, wallSize])
            cuboid(buttonBoxInsertSize, anchor=BOTTOM);
            
        move(sideWallMove)
        up(0.001)
        scale([grooveSlackMultipler,1,1])
            side_wall();
            
         move(frontWallMove)
        up(0.001)
        right(0.001)
      //  scale([1.02,1,1])
            front_wall(frontWallSize=frontWallGrooveSize);
        }
            
    }
    
    module side_wall(){

             
            cuboid(sideWallSize, anchor=BOT);
         
    }
	module modular_button_box(partType=partType){

        if(partType == "box" || partType == "all"){
            box();
        }
        if(partType == "frontWall" || partType == "all"){
        
        move(frontWallMove)
            front_wall(frontWallSize=frontWallSize, includeHole=true);

        }
       if(partType == "sideWall" || partType == "all"){
       
      move(sideWallMove)
            side_wall();
            
        }
	}


    sliced(renderType=renderType) {
        modular_button_box(partType=partType);
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

